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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/minio/madmin-go"

	"github.com/minio/minio/internal/logger"
	iampolicy "github.com/minio/pkg/iam/policy"
)

// SiteReplicationAdd - PUT /minio/admin/v3/site-replication/add
func (a adminAPIHandlers) SiteReplicationAdd(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "SiteReplicationAdd")

	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))

	objectAPI, cred := validateAdminReq(ctx, w, r, iampolicy.SiteReplicationAddAction)
	if objectAPI == nil {
		return
	}

	defer r.Body.Close()
	var addArg madmin.CRAdd
	errCode := readJSONBody(ctx, r.Body, &addArg, cred.SecretKey)
	if errCode != ErrNone {
		writeErrorResponseJSON(ctx, w, errorCodes.ToAPIErr(errCode), r.URL)
		return
	}

	err := globalClusterReplMgr.AddPeerClusters(ctx, addArg)
	fmt.Printf("AddPeerCluster err: %v\n", err)
	if err != nil {
		logger.LogIf(ctx, err)
		writeErrorResponseJSON(ctx, w, errorCodes.ToAPIErr(ErrInternalError), r.URL)
		return
	}
}

// CRInternalJoin - PUT /minio/admin/v3/site-replication/join
//
// used internally to tell current cluster to enable CR with
// the provided peer clusters and service account.
func (a adminAPIHandlers) CRInternalJoin(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "CRInternalJoin")

	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))

	objectAPI, cred := validateAdminReq(ctx, w, r, iampolicy.SiteReplicationAddAction)
	if objectAPI == nil {
		return
	}

	defer r.Body.Close()
	var joinArg madmin.CRInternalJoinReq
	errCode := readJSONBody(ctx, r.Body, &joinArg, cred.SecretKey)
	if errCode != ErrNone {
		writeErrorResponseJSON(ctx, w, errorCodes.ToAPIErr(errCode), r.URL)
		return
	}

	err := globalClusterReplMgr.InternalJoinReq(ctx, joinArg)
	if err != nil {
		logger.LogIf(ctx, err)
		writeErrorResponseJSON(ctx, w, errorCodes.ToAPIErr(ErrInternalError), r.URL)
		return
	}
}

// CRInternalBucketOps - PUT /minio/admin/v3/site-replication/bucket-ops?bucket=x&operation=y
func (a adminAPIHandlers) CRInternalBucketOps(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "CRInternalBucketOps")

	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))

	objectAPI, _ := validateAdminReq(ctx, w, r, iampolicy.SiteReplicationOperationAction)
	if objectAPI == nil {
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	operation := madmin.BktOp(vars["operation"])

	var err error
	switch operation {
	case madmin.MakeWithVersioningBktOp:
		_, isLockEnabled := r.Form["lockEnabled"]
		_, isVersioningEnabled := r.Form["versioningEnabled"]
		opts := BucketOptions{
			Location:          r.Form.Get("location"),
			LockEnabled:       isLockEnabled,
			VersioningEnabled: isVersioningEnabled,
		}
		err = globalClusterReplMgr.PeerBucketMakeWithVersioningHandler(ctx, bucket, opts)
	case madmin.ConfigureReplBktOp:
		err = globalClusterReplMgr.PeerBucketConfigureReplHandler(ctx, bucket)
	case madmin.DeleteBucketBktOp:
		err = globalClusterReplMgr.PeerBucketDeleteHandler(ctx, bucket, false)
	case madmin.ForceDeleteBucketBktOp:
		err = globalClusterReplMgr.PeerBucketDeleteHandler(ctx, bucket, true)
	default:
		writeErrorResponseJSON(ctx, w, errorCodes.ToAPIErr(ErrAdminInvalidArgument), r.URL)
		return
	}
	if err != nil {
		logger.LogIf(ctx, err)
		writeErrorResponseJSON(ctx, w, errorCodes.ToAPIErr(ErrInternalError), r.URL)
		return
	}

}

// CRInternalReplicateIAMItem - PUT /minio/admin/v3/site-replication/iam-item
func (a adminAPIHandlers) CRInternalReplicateIAMItem(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "CRInternalReplicateIAMItem")

	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))

	objectAPI, _ := validateAdminReq(ctx, w, r, iampolicy.SiteReplicationOperationAction)
	if objectAPI == nil {
		return
	}

	var item madmin.CRIAMItem
	errCode := readJSONBody(ctx, r.Body, &item, "")
	if errCode != ErrNone {
		writeErrorResponseJSON(ctx, w, errorCodes.ToAPIErr(errCode), r.URL)
		return
	}

	var err error
	switch item.Type {
	case madmin.CRIAMItemPolicy:
		err = globalClusterReplMgr.PeerAddPolicyHandler(ctx, item.Name, item.Policy)
	case madmin.CRIAMItemSvcAcc:
		err = globalClusterReplMgr.PeerSvcAccChangeHandler(ctx, *item.SvcAccChange)
	case madmin.CRIAMItemPolicyMapping:
		err = globalClusterReplMgr.PeerPolicyMappingHandler(ctx, *item.PolicyMapping)

	default:
		writeErrorResponseJSON(ctx, w, errorCodes.ToAPIErr(ErrAdminInvalidArgument), r.URL)
		return
	}
	if err != nil {
		logger.LogIf(ctx, err)
		writeErrorResponseJSON(ctx, w, errorCodes.ToAPIErr(ErrInternalError), r.URL)
		return
	}
}

// CRInternalReplicateBucketItem - PUT /minio/admin/v3/site-replication/bucket-meta
func (a adminAPIHandlers) CRInternalReplicateBucketItem(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "CRInternalReplicateIAMItem")

	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))

	objectAPI, _ := validateAdminReq(ctx, w, r, iampolicy.SiteReplicationOperationAction)
	if objectAPI == nil {
		return
	}

	var item madmin.CRBucketMeta
	errCode := readJSONBody(ctx, r.Body, &item, "")
	if errCode != ErrNone {
		writeErrorResponseJSON(ctx, w, errorCodes.ToAPIErr(errCode), r.URL)
		return
	}

	var err error
	switch item.Type {
	case madmin.CRBucketMetaTypePolicy:
		err = globalClusterReplMgr.PeerBucketPolicyHandler(ctx, item.Bucket, item.Policy)
	case madmin.CRBucketMetaTypeTags:
		err = globalClusterReplMgr.PeerBucketTaggingHandler(ctx, item.Bucket, item.Tags)
	case madmin.CRBucketMetaTypeObjectLockConfig:
		err = globalClusterReplMgr.PeerBucketObjectLockConfigHandler(ctx, item.Bucket, item.ObjectLockConfig)
	case madmin.CRBucketMetaTypeSSEConfig:
		err = globalClusterReplMgr.PeerBucketSSEConfigHandler(ctx, item.Bucket, item.SSEConfig)

	default:
		writeErrorResponseJSON(ctx, w, errorCodes.ToAPIErr(ErrAdminInvalidArgument), r.URL)
		return
	}
	if err != nil {
		logger.LogIf(ctx, err)
		writeErrorResponseJSON(ctx, w, errorCodes.ToAPIErr(ErrInternalError), r.URL)
		return
	}
}

// SiteReplicationDisable - PUT /minio/admin/v3/site-replication/disable
func (a adminAPIHandlers) SiteReplicationDisable(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "SiteReplicationDisable")

	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))

	objectAPI, _ := validateAdminReq(ctx, w, r, iampolicy.SiteReplicationDisableAction)
	if objectAPI == nil {
		return
	}

	return
}

// SiteReplicationInfo - GET /minio/admin/v3/site-replication/info
func (a adminAPIHandlers) SiteReplicationInfo(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "SiteReplicationInfo")

	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))

	objectAPI, _ := validateAdminReq(ctx, w, r, iampolicy.SiteReplicationInfoAction)
	if objectAPI == nil {
		return
	}

	info, err := globalClusterReplMgr.GetClusterInfo(ctx)
	if err != nil {
		writeErrorResponseJSON(ctx, w, toAdminAPIErr(ctx, err), r.URL)
		return
	}

	if err = json.NewEncoder(w).Encode(info); err != nil {
		writeErrorResponseJSON(ctx, w, toAdminAPIErr(ctx, err), r.URL)
		return
	}

	w.(http.Flusher).Flush()
}

func readJSONBody(ctx context.Context, body io.Reader, v interface{}, encryptionKey string) APIErrorCode {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return ErrInvalidRequest
	}

	if encryptionKey != "" {
		data, err = madmin.DecryptData(encryptionKey, bytes.NewReader(data))
		if err != nil {
			logger.LogIf(ctx, err)
			return ErrInvalidRequest
		}
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		return ErrAdminConfigBadJSON
	}

	return ErrNone
}
