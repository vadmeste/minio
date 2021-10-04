// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/minio/madmin-go"
	minioClient "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/replication"
	"github.com/minio/minio-go/v7/pkg/set"
	"github.com/minio/minio/internal/auth"
	sreplication "github.com/minio/minio/internal/bucket/replication"
	"github.com/minio/minio/internal/bucket/versioning"
	"github.com/minio/minio/internal/logger"
	"github.com/minio/pkg/bucket/policy"
	iampolicy "github.com/minio/pkg/iam/policy"
)

const (
	crStatePrefix = minioConfigPrefix + "/cluster-replication"

	crStateFile = "state.json"
)

const (
	crStateFormatVersion1 = 1
)

var (
	errCRCannotJoin   = errors.New("this cluster cannot be added to a CR set.")
	errCRSelfNotFound = errors.New("none of the given clusters correspond to the current one.")
	errCRPeerNotFound = errors.New("peer not found")
	errCRNotEnabled   = errors.New("cluster replication is not enabled")
)

func getCRStateFilePath() string {
	return crStatePrefix + SlashSeparator + crStateFile
}

// CRError - wrapped error for cluster replication.
type CRError struct {
	Cause error
	Msg   string
}

func (c CRError) Error() string {
	if c.Msg == "" {
		return fmt.Sprintf("ClusterReplication err: %v", c.Cause)
	}
	return fmt.Sprintf("ClusterReplication error: %s, cause: %v", c.Msg, c.Cause)
}

func wrapCRErr(err error) CRError {
	return CRError{Cause: err}
}

// ClusterReplMgr - manages cluster-level replication.
type ClusterReplMgr struct {
	sync.RWMutex

	enabled bool

	// In-memory and persisted multi-cluster replication state.
	state crState
}

type crState crStateV1

// crStateV1 represents version 1 of the cluster replication state persistence
// format.
type crStateV1 struct {
	Name string `json:"name"`

	// Peers maps peers by their deploymentID
	Peers                   map[string]madmin.PeerInfo `json:"peers"`
	ServiceAccountAccessKey string                     `json:"serviceAccountAccessKey"`
}

// crStateData represents the format of the current `crStateFile`.
type crStateData struct {
	Version int `json:"version"`

	CRState crStateV1 `json:"crState"`
}

// Init - initialize the cluster replication manager.
func (c *ClusterReplMgr) Init(ctx context.Context, objAPI ObjectLayer) error {
	err := c.loadFromDisk(ctx, objAPI)
	if err == errConfigNotFound {
		return nil
	}
	if c.enabled {
		fmt.Println("Cluster Replication is enabled.")
	}
	return err
}

func (c *ClusterReplMgr) loadFromDisk(ctx context.Context, objAPI ObjectLayer) error {
	buf, err := readConfig(ctx, objAPI, getCRStateFilePath())
	if err != nil {
		return err
	}

	// attempt to read just the version key in the state file to ensure we
	// are reading a compatible version.
	var ver struct {
		Version int `json:"version"`
	}
	err = json.Unmarshal(buf, &ver)
	if err != nil {
		return err
	}
	if ver.Version != crStateFormatVersion1 {
		return fmt.Errorf("Unexpected ClusterRepl state version: %d", ver.Version)
	}

	var sdata crStateData
	err = json.Unmarshal(buf, &sdata)
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()
	c.state = crState(sdata.CRState)
	c.enabled = true
	return nil
}

func (c *ClusterReplMgr) saveToDisk(ctx context.Context, state crState) error {
	sdata := crStateData{
		Version: crStateFormatVersion1,
		CRState: crStateV1(state),
	}
	buf, err := json.Marshal(sdata)
	if err != nil {
		return err
	}

	objAPI := newObjectLayerFn()
	if objAPI == nil {
		return errServerNotInitialized
	}
	err = saveConfig(ctx, objAPI, getCRStateFilePath(), buf)
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()
	c.state = state
	c.enabled = true
	return nil
}

const (
	// Access key of service account used for perform cluster-replication
	// operations.
	clusterReplicatorSvcAcc = "cluster-replicator-0"
)

// AddPeerClusters - add clusters for replication configuration.
func (c *ClusterReplMgr) AddPeerClusters(ctx context.Context, arg madmin.CRAdd) error {
	// If current cluster is already CR enabled, we fail.
	if c.enabled {
		return errCRCannotJoin
	}

	// Only one of the clusters being added, can have any buckets (i.e. self
	// here) - others must be empty.
	selfIdx := -1
	localHasBuckets := false
	nonLocalPeerWithBuckets := ""
	deploymentIDs := make([]string, 0, len(arg.Clusters))
	deploymentIDsSet := set.NewStringSet()
	for i, v := range arg.Clusters {
		admClient, err := getAdminClient(v.Endpoint, v.AccessKey, v.SecretKey)
		if err != nil {
			return fmt.Errorf("unable to create admin client: %w", err)
		}
		info, err := admClient.ServerInfo(ctx)
		if err != nil {
			return err
		}

		deploymentID := info.DeploymentID
		if deploymentID == "" {
			return fmt.Errorf("Cluster %s has an empty deploymentID", v.Name)
		}

		deploymentIDs = append(deploymentIDs, deploymentID)

		// deploymentIDs must be unique
		if deploymentIDsSet.Contains(deploymentID) {
			return fmt.Errorf("Found more than one cluster with the same deployment ID: %s", deploymentID)
		}
		deploymentIDsSet.Add(deploymentID)

		if deploymentID == globalDeploymentID {
			selfIdx = i
			objAPI := newObjectLayerFn()
			if objAPI == nil {
				return errServerNotInitialized
			}
			res, err := objAPI.ListBuckets(ctx)
			if err != nil {
				return err
			}
			if len(res) > 0 {
				localHasBuckets = true
			}
			continue
		}

		s3Client, err := getS3Client(v)
		if err != nil {
			return fmt.Errorf("unable to create s3 client: %w", err)
		}
		buckets, err := s3Client.ListBuckets(ctx)
		if err != nil {
			return err
		}

		if len(buckets) > 0 {
			nonLocalPeerWithBuckets = v.Name
		}
	}

	// For this `add` API, either all clusters must be empty or the local
	// cluster must be the only one having some buckets.

	if localHasBuckets && nonLocalPeerWithBuckets != "" {
		return fmt.Errorf("Only one cluster may have data when configuring site replication")
	}

	if !localHasBuckets && nonLocalPeerWithBuckets != "" {
		return fmt.Errorf("Please send your request to the cluster containing data/buckets: %s", nonLocalPeerWithBuckets)
	}

	// FIXME: Ideally, we also need to check if there are any global IAM
	// policies and any (LDAP user created) service accounts on the other
	// peer clusters , and if so, reject the cluster replicate add request.
	// This is not yet implemented.

	// FIXME: validate that all clusters are using the same (LDAP based)
	// external IDP.

	// VALIDATIONS COMPLETE.

	// Create a common service account for all clusters, with root
	// permissions.

	// Create a local service account.

	// Generate a secret key for the service account.
	var secretKey string
	{
		secretKeyBuf := make([]byte, 40)
		n, err := rand.Read(secretKeyBuf)
		if err == nil && n != 40 {
			err = fmt.Errorf("Unable to read 40 random bytes to generate secret key")
		}
		if err != nil {
			return wrapCRErr(err)
		}
		secretKey = strings.Replace(string([]byte(base64.StdEncoding.EncodeToString(secretKeyBuf))[:40]),
			"/", "+", -1)
	}

	svcCred, err := globalIAMSys.NewServiceAccount(ctx, arg.Clusters[selfIdx].AccessKey, nil, newServiceAccountOpts{
		accessKey: clusterReplicatorSvcAcc,
		secretKey: secretKey,
	})
	if err != nil {
		return fmt.Errorf("unable to create local service account: %w", err)
	}

	joinReq := madmin.CRInternalJoinReq{
		SvcAcctAccessKey: svcCred.AccessKey,
		SvcAcctSecretKey: svcCred.SecretKey,
		Peers:            make(map[string]madmin.PeerInfo),
	}
	for i, v := range arg.Clusters {
		joinReq.Peers[deploymentIDs[i]] = madmin.PeerInfo{
			Endpoint:     v.Endpoint,
			Name:         v.Name,
			DeploymentID: deploymentIDs[i],
		}
	}

	for i, v := range arg.Clusters {
		if i == selfIdx {
			continue
		}
		admClient, err := getAdminClient(v.Endpoint, v.AccessKey, v.SecretKey)
		if err != nil {
			return fmt.Errorf("unable to create admin client: %w", err)
		}
		joinReq.SvcAcctParent = v.AccessKey
		err = admClient.CRInternalJoin(ctx, joinReq)
		if err != nil {
			return fmt.Errorf("error in CR join: %v", err)
		}
	}

	// Other than handling existing buckets, we can now save the cluster
	// replication configuration state.
	state := crState{
		Name:                    arg.Clusters[selfIdx].Name,
		Peers:                   joinReq.Peers,
		ServiceAccountAccessKey: svcCred.AccessKey,
	}
	c.Lock()
	c.state = state
	c.enabled = true
	c.Unlock()

	return c.syncLocalToPeers(ctx)
}

// InternalJoinReq - internal API handler to respond to a peer cluster's request
// to join.
func (c *ClusterReplMgr) InternalJoinReq(ctx context.Context, arg madmin.CRInternalJoinReq) error {
	if c.enabled {
		return errCRCannotJoin
	}

	var ourName string
	for d, p := range arg.Peers {
		if d == globalDeploymentID {
			ourName = p.Name
			break
		}
	}
	if ourName == "" {
		return errCRSelfNotFound
	}

	_, err := globalIAMSys.NewServiceAccount(ctx, arg.SvcAcctParent, nil, newServiceAccountOpts{
		accessKey: arg.SvcAcctAccessKey,
		secretKey: arg.SvcAcctSecretKey,
	})
	if err != nil {
		return err
	}

	state := crState{
		Name:                    ourName,
		Peers:                   arg.Peers,
		ServiceAccountAccessKey: arg.SvcAcctAccessKey,
	}
	return c.saveToDisk(ctx, state)
}

// GetClusterInfo - returns cluster replication information.
func (c *ClusterReplMgr) GetClusterInfo(ctx context.Context) (info madmin.ClusterReplicateInfo, err error) {
	c.RLock()
	defer c.RUnlock()
	if !c.enabled {
		return info, nil
	}

	info.Enabled = true
	info.Name = c.state.Name
	info.Clusters = make([]madmin.PeerInfo, 0, len(c.state.Peers))
	for _, peer := range c.state.Peers {
		info.Clusters = append(info.Clusters, peer)
	}
	sort.SliceStable(info.Clusters, func(i, j int) bool {
		return info.Clusters[i].Name < info.Clusters[j].Name
	})

	info.ServiceAccountAccessKey = c.state.ServiceAccountAccessKey
	return info, nil
}

// MakeBucketHook - called during a regular make bucket call when cluster
// replication is enabled. It is responsible for the creation of the same bucket
// on remote clusters, and creating replication rules on local and peer
// clusters.
func (c *ClusterReplMgr) MakeBucketHook(ctx context.Context, bucket string, opts BucketOptions) error {
	// At this point, the local bucket is created.

	c.RLock()
	defer c.RUnlock()
	if !c.enabled {
		return nil
	}

	optsMap := make(map[string]string)
	if opts.Location != "" {
		optsMap["location"] = opts.Location
	}
	if opts.LockEnabled {
		optsMap["lockEnabled"] = ""
	}

	// Create bucket and enable versioning on all peers.
	makeBucketConcErr := c.concDo(
		func() error {
			err := c.PeerBucketMakeWithVersioningHandler(ctx, bucket, opts)
			if err != nil {
				fmt.Printf("%s: MakeWithVersioning: %v\n", c.state.Name, err)
			}
			return err
		},
		func(deploymentID string, p madmin.PeerInfo) error {
			admClient, err := c.getAdminClient(ctx, deploymentID)
			if err != nil {
				return err
			}

			err = admClient.CRInternalBucketOps(ctx, bucket, madmin.MakeWithVersioningBktOp, optsMap)
			if err != nil {
				fmt.Printf("%s->%s: MakeWithVersioning: %v\n", c.state.Name, p.Name, err)
			}
			return err
		},
	)
	// If all make-bucket-and-enable-versioning operations failed, nothing
	// more to do.
	if makeBucketConcErr.allFailed() {
		return makeBucketConcErr
	}

	// Log any errors in make-bucket operations.
	logger.LogIf(ctx, makeBucketConcErr.summaryErr)

	// Create bucket remotes and add replication rules for the bucket on
	// self and peers.
	makeRemotesConcErr := c.concDo(
		func() error {
			err := c.PeerBucketConfigureReplHandler(ctx, bucket)
			if err != nil {
				fmt.Printf("%s: ConfigureRepl: %v\n", c.state.Name, err)
			}
			return err
		},
		func(deploymentID string, p madmin.PeerInfo) error {
			admClient, err := c.getAdminClient(ctx, deploymentID)
			if err != nil {
				return err
			}

			err = admClient.CRInternalBucketOps(ctx, bucket, madmin.ConfigureReplBktOp, nil)
			if err != nil {
				fmt.Printf("%s->%s: ConfigureRepl: %v\n", c.state.Name, p.Name, err)
			}
			return err
		},
	)
	err := makeRemotesConcErr.summaryErr
	if err != nil {
		return err
	}

	return nil
}

// DeleteBucketHook - called during a regular delete bucket call when cluster
// replication is enabled. It is responsible for the deletion of the same bucket
// on remote clusters.
func (c *ClusterReplMgr) DeleteBucketHook(ctx context.Context, bucket string, forceDelete bool) error {
	// At this point, the local bucket is deleted.

	c.RLock()
	defer c.RUnlock()
	if !c.enabled {
		return nil
	}

	op := madmin.DeleteBucketBktOp
	if forceDelete {
		op = madmin.ForceDeleteBucketBktOp
	}

	// Send bucket delete to other clusters.
	cErr := c.concDo(nil, func(deploymentID string, p madmin.PeerInfo) error {
		admClient, err := c.getAdminClient(ctx, deploymentID)
		if err != nil {
			return wrapCRErr(err)
		}

		err = admClient.CRInternalBucketOps(ctx, bucket, op, nil)
		if err != nil {
			fmt.Printf("%s->%s: DeleteBucket: %v\n", c.state.Name, p.Name, err)
		}
		return err
	})
	return cErr.summaryErr
}

// PeerBucketMakeWithVersioningHandler - creates bucket and enables versioning.
func (c *ClusterReplMgr) PeerBucketMakeWithVersioningHandler(ctx context.Context, bucket string, opts BucketOptions) error {
	objAPI := newObjectLayerFn()
	if objAPI == nil {
		return errServerNotInitialized
	}
	err := objAPI.MakeBucketWithLocation(ctx, bucket, opts)
	if err != nil {
		// Check if this is a bucket exists error.
		_, ok1 := err.(BucketExists)
		_, ok2 := err.(BucketAlreadyExists)
		if !ok1 && !ok2 {
			fmt.Printf("%s: isFromPeer - mb err: %#v\n", c.state.Name, err)
			return wrapCRErr(err)
		}
	} else {
		// Load updated bucket metadata into memory as new
		// bucket was created.
		globalNotificationSys.LoadBucketMetadata(GlobalContext, bucket)
	}

	// Enable versioning on the bucket.
	config, err := globalBucketVersioningSys.Get(bucket)
	if err != nil {
		return wrapCRErr(err)
	}
	if !config.Enabled() {
		verConf := versioning.Versioning{
			Status: versioning.Enabled,
		}
		// FIXME: need to confirm if skipping object lock and
		// versioning-suspended state checks are valid here.
		cfgData, err := xml.Marshal(verConf)
		if err != nil {
			return wrapCRErr(err)
		}
		err = globalBucketMetadataSys.Update(bucket, bucketVersioningConfig, cfgData)
		if err != nil {
			fmt.Printf("%s: error enabling versioning: %#v\n", c.state.Name, err)
			return wrapCRErr(err)
		}
	}
	return nil
}

// PeerBucketConfigureReplHandler - configures replication remote and
// replication rules to all other peers for the local bucket.
func (c *ClusterReplMgr) PeerBucketConfigureReplHandler(ctx context.Context, bucket string) error {
	creds, err := c.getPeerCreds()
	if err != nil {
		return wrapCRErr(err)
	}

	// The following function, creates a bucket remote and sets up a bucket
	// replication rule for the given peer.
	configurePeerFn := func(d string, peer madmin.PeerInfo) error {
		ep, _ := url.Parse(peer.Endpoint)
		targets := globalBucketTargetSys.ListTargets(ctx, bucket, string(madmin.ReplicationService))
		targetARN := ""
		for _, target := range targets {
			if target.SourceBucket == bucket &&
				target.TargetBucket == bucket &&
				target.Endpoint == ep.Host &&
				target.Secure == (ep.Scheme == "https") &&
				target.Type == madmin.ReplicationService {
				targetARN = target.Arn
				break
			}
		}
		if targetARN == "" {
			bucketTarget := madmin.BucketTarget{
				SourceBucket: bucket,
				Endpoint:     ep.Host,
				Credentials: &madmin.Credentials{
					AccessKey: creds.AccessKey,
					SecretKey: creds.SecretKey,
				},
				TargetBucket:    bucket,
				Secure:          ep.Scheme == "https",
				API:             "s3v4",
				Type:            madmin.ReplicationService,
				Region:          "",
				ReplicationSync: false,
			}
			bucketTarget.Arn = globalBucketTargetSys.getRemoteARN(bucket, &bucketTarget)
			err := globalBucketTargetSys.SetTarget(ctx, bucket, &bucketTarget, false)
			if err != nil {
				fmt.Printf("%s->%s: target err: %#v\n", c.state.Name, peer.Name, err)
				return err
			}
			targets, err := globalBucketTargetSys.ListBucketTargets(ctx, bucket)
			if err != nil {
				return err
			}
			tgtBytes, err := json.Marshal(&targets)
			if err != nil {
				return err
			}
			if err = globalBucketMetadataSys.Update(bucket, bucketTargetsFile, tgtBytes); err != nil {
				return err
			}
			targetARN = bucketTarget.Arn
		}

		// Create bucket replication rule to this peer.

		// To add the bucket replication rule, we fetch the current
		// server configuration, and convert it to minio-go's
		// replication configuration type (by converting to xml and
		// parsing it back), use minio-go's add rule function, and
		// finally convert it back to the server type (again via xml).
		// This is needed as there is no add-rule function in the server
		// yet.

		// Though we do not check if the rule already exists, this is
		// not a problem as we are always using the same replication
		// rule ID - if the rule already exists, it is just replaced.
		replicationConfigS, err := globalBucketMetadataSys.GetReplicationConfig(ctx, bucket)
		if err != nil {
			_, ok := err.(BucketReplicationConfigNotFound)
			if !ok {
				return err
			}
		}
		var replicationConfig replication.Config
		if replicationConfigS != nil {
			replCfgSBytes, err := xml.Marshal(replicationConfigS)
			if err != nil {
				return err
			}
			err = xml.Unmarshal(replCfgSBytes, &replicationConfig)
			if err != nil {
				return err
			}
		}
		err = replicationConfig.AddRule(replication.Options{
			// Set the ID so we can identify the rule as being
			// created for site-replication and include the
			// destination cluster's deployment ID.
			ID: fmt.Sprintf("site-repl-%s", d),

			// Use a helper to generate unique priority numbers.
			Priority: fmt.Sprintf("%d", getPriorityHelper(replicationConfig)),

			Op:         replication.AddOption,
			RuleStatus: "enable",
			DestBucket: targetARN,

			// Replicate everything!
			ReplicateDeletes:        "enable",
			ReplicateDeleteMarkers:  "enable",
			ReplicaSync:             "enable",
			ExistingObjectReplicate: "enable",
		})
		if err != nil {
			fmt.Printf("%s->%s: rule add err: %#v\n", c.state.Name, peer.Name, err)
			return err
		}
		// Now convert the configuration back to server's type so we can
		// do some validation.
		newReplCfgBytes, err := xml.Marshal(replicationConfig)
		if err != nil {
			return err
		}
		newReplicationConfig, err := sreplication.ParseConfig(bytes.NewReader(newReplCfgBytes))
		if err != nil {
			return err
		}
		sameTarget, apiErr := validateReplicationDestination(ctx, bucket, newReplicationConfig)
		if apiErr != noError {
			return fmt.Errorf("bucket replication config validation error: %#v", apiErr)
		}
		err = newReplicationConfig.Validate(bucket, sameTarget)
		if err != nil {
			return err
		}
		// Config looks good, so we save it.
		replCfgData, err := xml.Marshal(newReplicationConfig)
		if err != nil {
			return err
		}
		err = globalBucketMetadataSys.Update(bucket, bucketReplicationConfig, replCfgData)
		if err != nil {
			fmt.Printf("%s->%s: replconferr: %v\n", c.state.Name, peer.Name, err)
		}
		return err
	}

	errMap := make(map[string]error, len(c.state.Peers))
	for d, peer := range c.state.Peers {
		if d == globalDeploymentID {
			continue
		}
		if err := configurePeerFn(d, peer); err != nil {
			errMap[d] = err
		}
	}
	return c.toErrorFromErrMap(errMap)
}

// PeerBucketDeleteHandler - deletes bucket on local in response to a delete
// bucket request from a peer.
func (c *ClusterReplMgr) PeerBucketDeleteHandler(ctx context.Context, bucket string, forceDelete bool) error {
	c.RLock()
	defer c.RUnlock()
	if !c.enabled {
		return errCRNotEnabled
	}

	// FIXME: need to handle cases where globalDNSConfig is set.

	objAPI := newObjectLayerFn()
	if objAPI == nil {
		return errServerNotInitialized
	}
	err := objAPI.DeleteBucket(ctx, bucket, forceDelete)
	if err != nil {
		return err
	}

	globalNotificationSys.DeleteBucketMetadata(ctx, bucket)

	return nil
}

// IAMChangeHook - called when IAM items need to be replicated to peer clusters.
// This includes named policy creation, policy mapping changes and service
// account changes.
//
// All policies are replicated.
//
// Policy mappings are only replicated when they are for LDAP users or groups
// (as an external IDP is always assumed when CR is used). In the case of
// OpenID, such mappings are provided from the IDP directly and so are not
// applicable here.
//
// Only certain service accounts can be replicated:
//
// Service accounts created for STS credentials using an external IDP: such STS
// credentials would be valid on the peer clusters as they are assumed to be
// using the same external IDP. Service accounts when using internal IDP or for
// root user will not be replicated.
func (c *ClusterReplMgr) IAMChangeHook(ctx context.Context, item madmin.CRIAMItem) error {
	// The IAM item has already been applied to the local cluster at this
	// point, and only needs to be updated on all remote peer clusters.

	c.RLock()
	defer c.RUnlock()
	if !c.enabled {
		return nil
	}

	cErr := c.concDo(nil, func(d string, p madmin.PeerInfo) error {
		admClient, err := c.getAdminClient(ctx, d)
		if err != nil {
			return wrapCRErr(err)
		}

		err = admClient.CRInternalReplicateIAMItem(ctx, item)
		if err != nil {
			fmt.Printf("%s->%s: CRInternalReplicateIAMItem: %v\n", c.state.Name, p.Name, err)
		}
		return err
	})
	return cErr.summaryErr
}

// PeerAddPolicyHandler - copies IAM policy to local. A nil policy argument,
// causes the named policy to be deleted.
func (c *ClusterReplMgr) PeerAddPolicyHandler(ctx context.Context, policyName string, p *iampolicy.Policy) error {
	var err error
	if p == nil {
		err = globalIAMSys.DeletePolicy(policyName)
	} else {
		err = globalIAMSys.SetPolicy(policyName, *p)
	}
	if err != nil {
		return wrapCRErr(err)
	}

	if p != nil {
		// Notify all other MinIO peers to reload policy
		for _, nerr := range globalNotificationSys.LoadPolicy(policyName) {
			if nerr.Err != nil {
				logger.GetReqInfo(ctx).SetTags("peerAddress", nerr.Host.String())
				logger.LogIf(ctx, nerr.Err)
			}
		}
		return nil
	}

	// Notify all other MinIO peers to delete policy
	for _, nerr := range globalNotificationSys.DeletePolicy(policyName) {
		if nerr.Err != nil {
			logger.GetReqInfo(ctx).SetTags("peerAddress", nerr.Host.String())
			logger.LogIf(ctx, nerr.Err)
		}
	}
	return nil
}

// PeerSvcAccChangeHandler - copies service-account change to local.
func (c *ClusterReplMgr) PeerSvcAccChangeHandler(ctx context.Context, change madmin.CRSvcAccChange) error {
	switch {
	case change.Create != nil:
		opts := newServiceAccountOpts{
			accessKey:     change.Create.AccessKey,
			secretKey:     change.Create.SecretKey,
			sessionPolicy: change.Create.SessionPolicy,
			ldapUsername:  change.Create.LDAPUser,
		}
		newCred, err := globalIAMSys.NewServiceAccount(ctx, change.Create.Parent, change.Create.Groups, opts)
		if err != nil {
			fmt.Printf("%s: NewServiceAccount: %#v\n", c.state.Name, err)
			return wrapCRErr(err)
		}

		// Notify all other Minio peers to reload user the service account
		for _, nerr := range globalNotificationSys.LoadServiceAccount(newCred.AccessKey) {
			if nerr.Err != nil {
				logger.GetReqInfo(ctx).SetTags("peerAddress", nerr.Host.String())
				logger.LogIf(ctx, nerr.Err)
			}
		}
	case change.Update != nil:
		opts := updateServiceAccountOpts{
			secretKey:     change.Update.SecretKey,
			status:        change.Update.Status,
			sessionPolicy: change.Update.SessionPolicy,
		}

		err := globalIAMSys.UpdateServiceAccount(ctx, change.Update.AccessKey, opts)
		if err != nil {
			fmt.Printf("%s: UpdateServiceAccount: %#v\n", c.state.Name, err)
			return wrapCRErr(err)
		}

		// Notify all other Minio peers to reload user the service account
		for _, nerr := range globalNotificationSys.LoadServiceAccount(change.Update.AccessKey) {
			if nerr.Err != nil {
				logger.GetReqInfo(ctx).SetTags("peerAddress", nerr.Host.String())
				logger.LogIf(ctx, nerr.Err)
			}
		}

	case change.Delete != nil:
		err := globalIAMSys.DeleteServiceAccount(ctx, change.Delete.AccessKey)
		if err != nil {
			fmt.Printf("%s: DeleteServiceAccount: %#v\n", c.state.Name, err)
			return wrapCRErr(err)
		}

		for _, nerr := range globalNotificationSys.DeleteServiceAccount(change.Delete.AccessKey) {
			if nerr.Err != nil {
				logger.GetReqInfo(ctx).SetTags("peerAddress", nerr.Host.String())
				logger.LogIf(ctx, nerr.Err)
			}
		}

	}

	return nil
}

// PeerPolicyMappingHandler - copies policy mapping to local.
func (c *ClusterReplMgr) PeerPolicyMappingHandler(ctx context.Context, mapping madmin.CRPolicyMapping) error {
	err := globalIAMSys.PolicyDBSet(mapping.UserOrGroup, mapping.Policy, mapping.IsGroup)
	if err != nil {
		return wrapCRErr(err)
	}

	// Notify all other MinIO peers to reload policy
	for _, nerr := range globalNotificationSys.LoadPolicyMapping(mapping.UserOrGroup, mapping.IsGroup) {
		if nerr.Err != nil {
			logger.GetReqInfo(ctx).SetTags("peerAddress", nerr.Host.String())
			logger.LogIf(ctx, nerr.Err)
		}
	}

	return nil
}

// BucketMetaHook - called when bucket meta changes happen and need to be
// replicated to peer clusters.
func (c *ClusterReplMgr) BucketMetaHook(ctx context.Context, item madmin.CRBucketMeta) error {
	// The change has already been applied to the local cluster at this
	// point, and only needs to be updated on all remote peer clusters.

	c.RLock()
	defer c.RUnlock()
	if !c.enabled {
		return nil
	}

	cErr := c.concDo(nil, func(d string, p madmin.PeerInfo) error {
		admClient, err := c.getAdminClient(ctx, d)
		if err != nil {
			return wrapCRErr(err)
		}

		err = admClient.CRInternalReplicateBucketMeta(ctx, item)
		if err != nil {
			fmt.Printf("%s->%s: CRInternalReplicateBucketMeta: %v\n", c.state.Name, p.Name, err)
		}
		return err
	})
	return cErr.summaryErr
}

// PeerBucketPolicyHandler - copies/deletes policy to local cluster.
func (c *ClusterReplMgr) PeerBucketPolicyHandler(ctx context.Context, bucket string, policy *policy.Policy) error {
	if policy != nil {
		configData, err := json.Marshal(policy)
		if err != nil {
			return wrapCRErr(err)
		}

		err = globalBucketMetadataSys.Update(bucket, bucketPolicyConfig, configData)
		if err != nil {
			return wrapCRErr(err)
		}
		return nil
	}

	// Delete the bucket policy
	err := globalBucketMetadataSys.Update(bucket, bucketPolicyConfig, nil)
	if err != nil {
		return wrapCRErr(err)
	}

	return nil
}

// PeerBucketTaggingHandler - copies/deletes tags to local cluster.
func (c *ClusterReplMgr) PeerBucketTaggingHandler(ctx context.Context, bucket string, tags *string) error {
	if tags != nil {
		configData, err := base64.StdEncoding.DecodeString(*tags)
		if err != nil {
			return wrapCRErr(err)
		}
		err = globalBucketMetadataSys.Update(bucket, bucketTaggingConfig, configData)
		if err != nil {
			return wrapCRErr(err)
		}
		return nil
	}

	// Delete the tags
	err := globalBucketMetadataSys.Update(bucket, bucketTaggingConfig, nil)
	if err != nil {
		return wrapCRErr(err)
	}

	return nil
}

// PeerBucketObjectLockConfigHandler - sets object lock on local bucket.
func (c *ClusterReplMgr) PeerBucketObjectLockConfigHandler(ctx context.Context, bucket string, objectLockData *string) error {
	if objectLockData != nil {
		configData, err := base64.StdEncoding.DecodeString(*objectLockData)
		if err != nil {
			return wrapCRErr(err)
		}
		err = globalBucketMetadataSys.Update(bucket, objectLockConfig, configData)
		if err != nil {
			return wrapCRErr(err)
		}
		return nil
	}

	return nil
}

// PeerBucketSSEConfigHandler - copies/deletes SSE config to local cluster.
func (c *ClusterReplMgr) PeerBucketSSEConfigHandler(ctx context.Context, bucket string, sseConfig *string) error {
	if sseConfig != nil {
		configData, err := base64.StdEncoding.DecodeString(*sseConfig)
		if err != nil {
			return wrapCRErr(err)
		}
		err = globalBucketMetadataSys.Update(bucket, bucketSSEConfig, configData)
		if err != nil {
			return wrapCRErr(err)
		}
		return nil
	}

	// Delete sse config
	err := globalBucketMetadataSys.Update(bucket, bucketSSEConfig, nil)
	if err != nil {
		return wrapCRErr(err)
	}
	return nil
}

// getAdminClient - NOTE: ensure to take at least a read lock on ClusterReplMgr
// before calling this.
func (c *ClusterReplMgr) getAdminClient(ctx context.Context, deploymentID string) (*madmin.AdminClient, error) {
	creds, err := c.getPeerCreds()
	if err != nil {
		return nil, err
	}

	peer, ok := c.state.Peers[deploymentID]
	if !ok {
		return nil, errCRPeerNotFound
	}

	return getAdminClient(peer.Endpoint, creds.AccessKey, creds.SecretKey)
}

func (c *ClusterReplMgr) getPeerCreds() (*auth.Credentials, error) {
	globalIAMSys.store.rlock()
	defer globalIAMSys.store.runlock()
	creds, ok := globalIAMSys.iamUsersMap[c.state.ServiceAccountAccessKey]
	if !ok {
		return nil, errors.New("cluster replication service account not found!")
	}
	return &creds, nil
}

// syncLocalToPeers is used when initially configuring cluster replication, to
// copy existing buckets, their settings, service accounts and policies to all
// new peers.
func (c *ClusterReplMgr) syncLocalToPeers(ctx context.Context) error {

	// If local has buckets, enable versioning on them, create them on peers
	// and setup replication rules.
	objAPI := newObjectLayerFn()
	if objAPI == nil {
		return errServerNotInitialized
	}
	buckets, err := objAPI.ListBuckets(ctx)
	if err != nil {
		return err
	}
	for _, bucketInfo := range buckets {
		bucket := bucketInfo.Name

		// MinIO does not store bucket location - so we just check if
		// object locking is enabled.
		lockConfig, err := globalBucketMetadataSys.GetObjectLockConfig(bucket)
		if err != nil {
			if _, ok := err.(BucketObjectLockConfigNotFound); !ok {
				return err
			}
		}
		opts := BucketOptions{
			LockEnabled: lockConfig.ObjectLockEnabled == "Enabled",
		}

		// Now call the MakeBucketHook on existing bucket - this will
		// create buckets and replication rules on peer clusters.
		err = c.MakeBucketHook(ctx, bucket, opts)
		if err != nil {
			return wrapCRErr(err)
		}

		// Replicate bucket policy if present.
		policy, err := globalPolicySys.Get(bucket)
		found := true
		if err != nil {
			if _, ok := err.(BucketPolicyNotFound); ok {
				found = false
			} else {
				return wrapCRErr(err)
			}
		}
		if found {
			err = c.BucketMetaHook(ctx, madmin.CRBucketMeta{
				Type:   madmin.CRBucketMetaTypePolicy,
				Bucket: bucket,
				Policy: policy,
			})
			if err != nil {
				return wrapCRErr(err)
			}
		}

		// Replicate bucket tags if present.
		tags, err := globalBucketMetadataSys.GetTaggingConfig(bucket)
		found = true
		if err != nil {
			if _, ok := err.(BucketTaggingNotFound); ok {
				found = false
			} else {
				return wrapCRErr(err)
			}
		}
		if found {
			tagCfg, err := xml.Marshal(tags)
			if err != nil {
				return wrapCRErr(err)
			}
			tagCfgStr := base64.StdEncoding.EncodeToString(tagCfg)
			err = c.BucketMetaHook(ctx, madmin.CRBucketMeta{
				Type:   madmin.CRBucketMetaTypeTags,
				Bucket: bucket,
				Tags:   &tagCfgStr,
			})
			if err != nil {
				return wrapCRErr(err)
			}
		}

		// Replicate object-lock config if present.
		objLockCfg, err := globalBucketMetadataSys.GetObjectLockConfig(bucket)
		found = true
		if err != nil {
			if _, ok := err.(BucketObjectLockConfigNotFound); ok {
				found = false
			} else {
				return wrapCRErr(err)
			}
		}
		if found {
			objLockCfgData, err := xml.Marshal(objLockCfg)
			if err != nil {
				return wrapCRErr(err)
			}
			objLockStr := base64.StdEncoding.EncodeToString(objLockCfgData)
			err = c.BucketMetaHook(ctx, madmin.CRBucketMeta{
				Type:   madmin.CRBucketMetaTypeObjectLockConfig,
				Bucket: bucket,
				Tags:   &objLockStr,
			})
			if err != nil {
				return wrapCRErr(err)
			}
		}

		// Replicate existing bucket bucket encryption settings
		sseConfig, err := globalBucketMetadataSys.GetSSEConfig(bucket)
		found = true
		if err != nil {
			if _, ok := err.(BucketSSEConfigNotFound); ok {
				found = false
			} else {
				return wrapCRErr(err)
			}
		}
		if found {
			sseConfigData, err := xml.Marshal(sseConfig)
			if err != nil {
				return wrapCRErr(err)
			}
			sseConfigStr := base64.StdEncoding.EncodeToString(sseConfigData)
			err = c.BucketMetaHook(ctx, madmin.CRBucketMeta{
				Type:      madmin.CRBucketMetaTypeSSEConfig,
				Bucket:    bucket,
				SSEConfig: &sseConfigStr,
			})
		}
	}

	{
		// Replicate IAM policies on local to all peers.
		allPolicies, err := globalIAMSys.ListPolicies("")
		if err != nil {
			return wrapCRErr(err)
		}

		for pname, policy := range allPolicies {
			err := c.IAMChangeHook(ctx, madmin.CRIAMItem{
				Type:   madmin.CRIAMItemPolicy,
				Name:   pname,
				Policy: &policy,
			})
			if err != nil {
				return wrapCRErr(err)
			}
		}
	}

	{
		// Replicate policy mappings on local to all peers.
		userPolicyMap := make(map[string]MappedPolicy)
		groupPolicyMap := make(map[string]MappedPolicy)
		globalIAMSys.store.rlock()
		errU := globalIAMSys.store.loadMappedPolicies(ctx, stsUser, false, userPolicyMap)
		errG := globalIAMSys.store.loadMappedPolicies(ctx, stsUser, true, groupPolicyMap)
		globalIAMSys.store.runlock()
		if errU != nil {
			return wrapCRErr(errU)
		}
		if errG != nil {
			return wrapCRErr(errG)
		}

		for user, mp := range userPolicyMap {
			err := c.IAMChangeHook(ctx, madmin.CRIAMItem{
				Type: madmin.CRIAMItemPolicyMapping,
				PolicyMapping: &madmin.CRPolicyMapping{
					UserOrGroup: user,
					IsGroup:     false,
					Policy:      mp.Policies,
				},
			})
			if err != nil {
				return wrapCRErr(err)
			}
		}

		for group, mp := range groupPolicyMap {
			err := c.IAMChangeHook(ctx, madmin.CRIAMItem{
				Type: madmin.CRIAMItemPolicyMapping,
				PolicyMapping: &madmin.CRPolicyMapping{
					UserOrGroup: group,
					IsGroup:     true,
					Policy:      mp.Policies,
				},
			})
			if err != nil {
				return wrapCRErr(err)
			}
		}
	}

	{
		// Check for service accounts and replicate them. Only LDAP user
		// owned service accounts are supported for this operation.
		serviceAccounts := make(map[string]auth.Credentials)
		globalIAMSys.store.rlock()
		err := globalIAMSys.store.loadUsers(ctx, svcUser, serviceAccounts)
		globalIAMSys.store.runlock()
		if err != nil {
			return wrapCRErr(err)
		}
		for user, acc := range serviceAccounts {
			ldapUser, err := globalIAMSys.GetLDAPUserForSvcAcc(ctx, acc.AccessKey)
			if err != nil {
				return wrapCRErr(err)
			}
			if ldapUser == "" {
				continue
			}
			_, policy, err := globalIAMSys.GetServiceAccount(ctx, acc.AccessKey)
			if err != nil {
				return wrapCRErr(err)
			}
			err = c.IAMChangeHook(ctx, madmin.CRIAMItem{
				Type: madmin.CRIAMItemSvcAcc,
				SvcAccChange: &madmin.CRSvcAccChange{
					Create: &madmin.CRSvcAccCreate{
						Parent:        acc.ParentUser,
						AccessKey:     user,
						SecretKey:     acc.SecretKey,
						Groups:        acc.Groups,
						LDAPUser:      ldapUser,
						SessionPolicy: policy,
						Status:        acc.Status,
					},
				},
			})
			if err != nil {
				return wrapCRErr(err)
			}
		}
	}

	return nil
}

// Concurrency helpers

type concErr struct {
	numActions int
	errMap     map[string]error
	summaryErr error
}

func (c concErr) Error() string {
	return c.summaryErr.Error()
}

func (c concErr) allFailed() bool {
	return len(c.errMap) == c.numActions
}

func (c *ClusterReplMgr) toErrorFromErrMap(errMap map[string]error) error {
	if len(errMap) == 0 {
		return nil
	}

	msgs := []string{}
	for d, err := range errMap {
		name := c.state.Peers[d].Name
		msgs = append(msgs, fmt.Sprintf("Site %s (%s): %v", name, d, err))
	}
	return fmt.Errorf("Site replication error(s): %s", strings.Join(msgs, "; "))
}

func (c *ClusterReplMgr) newConcErr(numActions int, errMap map[string]error) concErr {
	return concErr{
		numActions: numActions,
		errMap:     errMap,
		summaryErr: c.toErrorFromErrMap(errMap),
	}
}

// concDo calls actions concurrently. selfActionFn is run for the current
// cluster and peerActionFn is run for each peer replication cluster.
func (c *ClusterReplMgr) concDo(selfActionFn func() error, peerActionFn func(deploymentID string, p madmin.PeerInfo) error) concErr {
	depIDs := make([]string, 0, len(c.state.Peers))
	for d := range c.state.Peers {
		depIDs = append(depIDs, d)
	}
	errs := make([]error, len(c.state.Peers))
	var wg sync.WaitGroup
	for i := range depIDs {
		wg.Add(1)
		go func(i int) {
			if depIDs[i] == globalDeploymentID {
				if selfActionFn != nil {
					errs[i] = selfActionFn()
				}
			} else {
				errs[i] = peerActionFn(depIDs[i], c.state.Peers[depIDs[i]])
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	errMap := make(map[string]error, len(c.state.Peers))
	for i, depID := range depIDs {
		if errs[i] != nil {
			errMap[depID] = errs[i]
		}
	}
	numActions := len(c.state.Peers) - 1
	if selfActionFn != nil {
		numActions++
	}
	return c.newConcErr(numActions, errMap)
}

// Other helpers

// newRemoteClusterHTTPTransport returns a new http configuration
// used while communicating with the remote cluster.
func newRemoteClusterHTTPTransport() *http.Transport {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			RootCAs: globalRootCAs,
		},
	}
	return tr
}

func getAdminClient(endpoint, accessKey, secretKey string) (*madmin.AdminClient, error) {
	epURL, _ := url.Parse(endpoint)
	client, err := madmin.New(epURL.Host, accessKey, secretKey, epURL.Scheme == "https")
	if err != nil {
		return nil, err
	}
	client.SetCustomTransport(newRemoteClusterHTTPTransport())
	return client, nil
}

func getS3Client(pc madmin.PeerCluster) (*minioClient.Client, error) {
	ep, _ := url.Parse(pc.Endpoint)
	return minioClient.New(ep.Host, &minioClient.Options{
		Creds:     credentials.NewStaticV4(pc.AccessKey, pc.SecretKey, ""),
		Secure:    ep.Scheme == "https",
		Transport: newRemoteClusterHTTPTransport(),
	})
}

func getPriorityHelper(replicationConfig replication.Config) int {
	maxPrio := 0
	for _, rule := range replicationConfig.Rules {
		if rule.Priority > maxPrio {
			maxPrio = rule.Priority
		}
	}

	// leave some gaps in priority numbers for flexibility
	return maxPrio + 10
}
