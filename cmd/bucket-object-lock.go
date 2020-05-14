/*
 * MinIO Cloud Storage, (C) 2019-2020 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"math"
	"net/http"
	"path"
	"sync"

	"github.com/minio/minio/cmd/logger"
	"github.com/minio/minio/pkg/auth"
	objectlock "github.com/minio/minio/pkg/bucket/object/lock"
	"github.com/minio/minio/pkg/bucket/policy"
)

// BucketObjectLockSys - map of bucket and retention configuration.
type BucketObjectLockSys struct {
	sync.RWMutex
	retentionMap map[string]*objectlock.Retention
}

// Set - set retention configuration.
func (sys *BucketObjectLockSys) Set(bucketName string, retention *objectlock.Retention) {
	if globalIsGateway {
		// no-op
		return
	}

	sys.Lock()
	sys.retentionMap[bucketName] = retention
	sys.Unlock()
}

// Get - Get retention configuration.
func (sys *BucketObjectLockSys) Get(bucketName string) (r *objectlock.Retention, ok bool) {
	if globalIsGateway {
		// When gateway is enabled, no cached value
		// is used to validate bucket object lock configuration
		objAPI := newObjectLayerWithoutSafeModeFn()
		if objAPI == nil {
			return
		}

		lc, err := objAPI.GetBucketObjectLockConfig(GlobalContext, bucketName)
		if err != nil {
			return
		}
		return lc.ToRetention(), true
	}

	if sys == nil {
		return
	}

	sys.RLock()
	defer sys.RUnlock()
	r, ok = sys.retentionMap[bucketName]
	return r, ok
}

// Remove - removes retention sysuration.
func (sys *BucketObjectLockSys) Remove(bucketName string) {
	if sys == nil {
		return
	}
	sys.Lock()
	delete(sys.retentionMap, bucketName)
	sys.Unlock()
}

// Similar to enforceRetentionBypassForDelete but for WebUI
func enforceRetentionBypassForDeleteWeb(ctx context.Context, r *http.Request, bucket, object string, getObjectInfoFn GetObjectInfoFn, govBypassPerms bool) APIErrorCode {
	opts, err := getOpts(ctx, r, bucket, object)
	if err != nil {
		return toAPIErrorCode(ctx, err)
	}

	oi, err := getObjectInfoFn(ctx, bucket, object, opts)
	if err != nil {
		return toAPIErrorCode(ctx, err)
	}

	lhold := objectlock.GetObjectLegalHoldMeta(oi.UserDefined)
	if lhold.Status.Valid() && lhold.Status == objectlock.LegalHoldOn {
		return ErrObjectLocked
	}

	ret := objectlock.GetObjectRetentionMeta(oi.UserDefined)
	if ret.Mode.Valid() {
		switch ret.Mode {
		case objectlock.RetCompliance:
			// In compliance mode, a protected object version can't be overwritten
			// or deleted by any user, including the root user in your AWS account.
			// When an object is locked in compliance mode, its retention mode can't
			// be changed, and its retention period can't be shortened. Compliance mode
			// ensures that an object version can't be overwritten or deleted for the
			// duration of the retention period.
			t, err := objectlock.UTCNowNTP()
			if err != nil {
				logger.LogIf(ctx, err)
				return ErrObjectLocked
			}

			if !ret.RetainUntilDate.Before(t) {
				return ErrObjectLocked
			}
			return ErrNone
		case objectlock.RetGovernance:
			// In governance mode, users can't overwrite or delete an object
			// version or alter its lock settings unless they have special
			// permissions. With governance mode, you protect objects against
			// being deleted by most users, but you can still grant some users
			// permission to alter the retention settings or delete the object
			// if necessary. You can also use governance mode to test retention-period
			// settings before creating a compliance-mode retention period.
			// To override or remove governance-mode retention settings, a
			// user must have the s3:BypassGovernanceRetention permission
			// and must explicitly include x-amz-bypass-governance-retention:true
			// as a request header with any request that requires overriding
			// governance mode.
			byPassSet := govBypassPerms && objectlock.IsObjectLockGovernanceBypassSet(r.Header)
			if !byPassSet {
				t, err := objectlock.UTCNowNTP()
				if err != nil {
					logger.LogIf(ctx, err)
					return ErrObjectLocked
				}

				if !ret.RetainUntilDate.Before(t) {
					return ErrObjectLocked
				}

				if !govBypassPerms {
					return ErrObjectLocked
				}

				return ErrNone
			}
		}
	}
	return ErrNone
}

// enforceRetentionForDeletion checks if it is appropriate to remove an
// object according to locking configuration when this is lifecycle/ bucket quota asking.
func enforceRetentionForDeletion(ctx context.Context, objInfo ObjectInfo) (locked bool) {
	lhold := objectlock.GetObjectLegalHoldMeta(objInfo.UserDefined)
	if lhold.Status.Valid() && lhold.Status == objectlock.LegalHoldOn {
		return true
	}

	ret := objectlock.GetObjectRetentionMeta(objInfo.UserDefined)
	if ret.Mode.Valid() && (ret.Mode == objectlock.RetCompliance || ret.Mode == objectlock.RetGovernance) {
		t, err := objectlock.UTCNowNTP()
		if err != nil {
			logger.LogIf(ctx, err)
			return true
		}
		if ret.RetainUntilDate.After(t) {
			return true
		}
	}
	return false
}

// enforceRetentionBypassForDelete enforces whether an existing object under governance can be deleted
// with governance bypass headers set in the request.
// Objects under site wide WORM can never be overwritten.
// For objects in "Governance" mode, overwrite is allowed if a) object retention date is past OR
// governance bypass headers are set and user has governance bypass permissions.
// Objects in "Compliance" mode can be overwritten only if retention date is past.
func enforceRetentionBypassForDelete(ctx context.Context, r *http.Request, bucket, object string, getObjectInfoFn GetObjectInfoFn) APIErrorCode {
	opts, err := getOpts(ctx, r, bucket, object)
	if err != nil {
		return toAPIErrorCode(ctx, err)
	}

	oi, err := getObjectInfoFn(ctx, bucket, object, opts)
	if err != nil {
		return toAPIErrorCode(ctx, err)
	}

	lhold := objectlock.GetObjectLegalHoldMeta(oi.UserDefined)
	if lhold.Status.Valid() && lhold.Status == objectlock.LegalHoldOn {
		return ErrObjectLocked
	}

	ret := objectlock.GetObjectRetentionMeta(oi.UserDefined)
	if ret.Mode.Valid() {
		switch ret.Mode {
		case objectlock.RetCompliance:
			// In compliance mode, a protected object version can't be overwritten
			// or deleted by any user, including the root user in your AWS account.
			// When an object is locked in compliance mode, its retention mode can't
			// be changed, and its retention period can't be shortened. Compliance mode
			// ensures that an object version can't be overwritten or deleted for the
			// duration of the retention period.
			t, err := objectlock.UTCNowNTP()
			if err != nil {
				logger.LogIf(ctx, err)
				return ErrObjectLocked
			}

			if !ret.RetainUntilDate.Before(t) {
				return ErrObjectLocked
			}
			return ErrNone
		case objectlock.RetGovernance:
			// In governance mode, users can't overwrite or delete an object
			// version or alter its lock settings unless they have special
			// permissions. With governance mode, you protect objects against
			// being deleted by most users, but you can still grant some users
			// permission to alter the retention settings or delete the object
			// if necessary. You can also use governance mode to test retention-period
			// settings before creating a compliance-mode retention period.
			// To override or remove governance-mode retention settings, a
			// user must have the s3:BypassGovernanceRetention permission
			// and must explicitly include x-amz-bypass-governance-retention:true
			// as a request header with any request that requires overriding
			// governance mode.
			//
			byPassSet := objectlock.IsObjectLockGovernanceBypassSet(r.Header)
			if !byPassSet {
				t, err := objectlock.UTCNowNTP()
				if err != nil {
					logger.LogIf(ctx, err)
					return ErrObjectLocked
				}

				if !ret.RetainUntilDate.Before(t) {
					return ErrObjectLocked
				}
				return ErrNone
			}
			// https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lock-overview.html#object-lock-retention-modes
			// If you try to delete objects protected by governance mode and have s3:BypassGovernanceRetention
			// or s3:GetBucketObjectLockConfiguration permissions, the operation will succeed.
			govBypassPerms1 := checkRequestAuthType(ctx, r, policy.BypassGovernanceRetentionAction, bucket, object)
			govBypassPerms2 := checkRequestAuthType(ctx, r, policy.GetBucketObjectLockConfigurationAction, bucket, object)
			if govBypassPerms1 != ErrNone && govBypassPerms2 != ErrNone {
				return ErrAccessDenied
			}
		}
	}
	return ErrNone
}

// enforceRetentionBypassForPut enforces whether an existing object under governance can be overwritten
// with governance bypass headers set in the request.
// Objects under site wide WORM cannot be overwritten.
// For objects in "Governance" mode, overwrite is allowed if a) object retention date is past OR
// governance bypass headers are set and user has governance bypass permissions.
// Objects in compliance mode can be overwritten only if retention date is being extended. No mode change is permitted.
func enforceRetentionBypassForPut(ctx context.Context, r *http.Request, bucket, object string, getObjectInfoFn GetObjectInfoFn, objRetention *objectlock.ObjectRetention, cred auth.Credentials, owner bool, claims map[string]interface{}) (ObjectInfo, APIErrorCode) {
	byPassSet := objectlock.IsObjectLockGovernanceBypassSet(r.Header)
	opts, err := getOpts(ctx, r, bucket, object)
	if err != nil {
		return ObjectInfo{}, toAPIErrorCode(ctx, err)
	}

	oi, err := getObjectInfoFn(ctx, bucket, object, opts)
	if err != nil {
		return oi, toAPIErrorCode(ctx, err)
	}

	t, err := objectlock.UTCNowNTP()
	if err != nil {
		logger.LogIf(ctx, err)
		return oi, ErrObjectLocked
	}

	// Pass in relative days from current time, to additionally to verify "object-lock-remaining-retention-days" policy if any.
	days := int(math.Ceil(math.Abs(objRetention.RetainUntilDate.Sub(t).Hours()) / 24))

	ret := objectlock.GetObjectRetentionMeta(oi.UserDefined)
	if ret.Mode.Valid() {
		// Retention has expired you may change whatever you like.
		if ret.RetainUntilDate.Before(t) {
			perm := isPutRetentionAllowed(bucket, object,
				days, objRetention.RetainUntilDate.Time,
				objRetention.Mode, byPassSet, r, cred,
				owner, claims)
			return oi, perm
		}

		switch ret.Mode {
		case objectlock.RetGovernance:
			govPerm := isPutRetentionAllowed(bucket, object, days,
				objRetention.RetainUntilDate.Time, objRetention.Mode,
				byPassSet, r, cred, owner, claims)
			// Governance mode retention period cannot be shortened, if x-amz-bypass-governance is not set.
			if !byPassSet {
				if objRetention.Mode != objectlock.RetGovernance || objRetention.RetainUntilDate.Before((ret.RetainUntilDate.Time)) {
					return oi, ErrObjectLocked
				}
			}
			return oi, govPerm
		case objectlock.RetCompliance:
			// Compliance retention mode cannot be changed or shortened.
			// https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lock-overview.html#object-lock-retention-modes
			if objRetention.Mode != objectlock.RetCompliance || objRetention.RetainUntilDate.Before((ret.RetainUntilDate.Time)) {
				return oi, ErrObjectLocked
			}
			compliancePerm := isPutRetentionAllowed(bucket, object,
				days, objRetention.RetainUntilDate.Time, objRetention.Mode,
				false, r, cred, owner, claims)
			return oi, compliancePerm
		}
		return oi, ErrNone
	} // No pre-existing retention metadata present.

	perm := isPutRetentionAllowed(bucket, object,
		days, objRetention.RetainUntilDate.Time,
		objRetention.Mode, byPassSet, r, cred, owner, claims)
	return oi, perm
}

// checkPutObjectLockAllowed enforces object retention policy and legal hold policy
// for requests with WORM headers
// See https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lock-managing.html for the spec.
// For non-existing objects with object retention headers set, this method returns ErrNone if bucket has
// locking enabled and user has requisite permissions (s3:PutObjectRetention)
// If object exists on object store and site wide WORM enabled - this method
// returns an error. For objects in "Governance" mode, overwrite is allowed if the retention date has expired.
// For objects in "Compliance" mode, retention date cannot be shortened, and mode cannot be altered.
// For objects with legal hold header set, the s3:PutObjectLegalHold permission is expected to be set
// Both legal hold and retention can be applied independently on an object
func checkPutObjectLockAllowed(ctx context.Context, r *http.Request, bucket, object string, getObjectInfoFn GetObjectInfoFn, retentionPermErr, legalHoldPermErr APIErrorCode) (objectlock.RetMode, objectlock.RetentionDate, objectlock.ObjectLegalHold, APIErrorCode) {
	var mode objectlock.RetMode
	var retainDate objectlock.RetentionDate
	var legalHold objectlock.ObjectLegalHold

	retentionRequested := objectlock.IsObjectLockRetentionRequested(r.Header)
	legalHoldRequested := objectlock.IsObjectLockLegalHoldRequested(r.Header)

	retentionCfg, isWORMBucket := globalBucketObjectLockSys.Get(bucket)
	if !isWORMBucket {
		if legalHoldRequested || retentionRequested {
			return mode, retainDate, legalHold, ErrInvalidBucketObjectLockConfiguration
		}
		// If this not a WORM enabled bucket, we should return right here.
		return mode, retainDate, legalHold, ErrNone
	}

	var objExists bool
	opts, err := getOpts(ctx, r, bucket, object)
	if err != nil {
		return mode, retainDate, legalHold, toAPIErrorCode(ctx, err)
	}

	t, err := objectlock.UTCNowNTP()
	if err != nil {
		logger.LogIf(ctx, err)
		return mode, retainDate, legalHold, ErrObjectLocked
	}

	if objInfo, err := getObjectInfoFn(ctx, bucket, object, opts); err == nil {
		objExists = true
		r := objectlock.GetObjectRetentionMeta(objInfo.UserDefined)
		if r.Mode == objectlock.RetCompliance && r.RetainUntilDate.After(t) {
			return mode, retainDate, legalHold, ErrObjectLocked
		}
		mode = r.Mode
		retainDate = r.RetainUntilDate
		legalHold = objectlock.GetObjectLegalHoldMeta(objInfo.UserDefined)
		// Disallow overwriting an object on legal hold
		if legalHold.Status == objectlock.LegalHoldOn {
			return mode, retainDate, legalHold, ErrObjectLocked
		}
	}

	if legalHoldRequested {
		var lerr error
		if legalHold, lerr = objectlock.ParseObjectLockLegalHoldHeaders(r.Header); lerr != nil {
			return mode, retainDate, legalHold, toAPIErrorCode(ctx, err)
		}
	}

	if retentionRequested {
		legalHold, err := objectlock.ParseObjectLockLegalHoldHeaders(r.Header)
		if err != nil {
			return mode, retainDate, legalHold, toAPIErrorCode(ctx, err)
		}
		rMode, rDate, err := objectlock.ParseObjectLockRetentionHeaders(r.Header)
		if err != nil {
			return mode, retainDate, legalHold, toAPIErrorCode(ctx, err)
		}
		if objExists && retainDate.After(t) {
			return mode, retainDate, legalHold, ErrObjectLocked
		}
		if retentionPermErr != ErrNone {
			return mode, retainDate, legalHold, retentionPermErr
		}
		return rMode, rDate, legalHold, ErrNone
	}

	if !retentionRequested && isWORMBucket {
		if retentionPermErr != ErrNone {
			return mode, retainDate, legalHold, retentionPermErr
		}
		t, err := objectlock.UTCNowNTP()
		if err != nil {
			logger.LogIf(ctx, err)
			return mode, retainDate, legalHold, ErrObjectLocked
		}
		// AWS S3 just creates a new version of object when an object is being overwritten.
		if objExists && retainDate.After(t) {
			return mode, retainDate, legalHold, ErrObjectLocked
		}
		if !legalHoldRequested && !retentionCfg.IsEmpty() {
			// inherit retention from bucket configuration
			return retentionCfg.Mode, objectlock.RetentionDate{Time: t.Add(retentionCfg.Validity)}, legalHold, ErrNone
		}
		return "", objectlock.RetentionDate{}, legalHold, ErrNone
	}
	return mode, retainDate, legalHold, ErrNone
}

func readBucketObjectLockConfig(ctx context.Context, objAPI ObjectLayer, bucket string) (*objectlock.Config, error) {
	meta, err := loadBucketMetadata(ctx, objAPI, bucket)
	if err != nil && err != errMetaDataConverted {
		return nil, toObjectErr(err, bucket)
	}
	if !meta.LockEnabled {
		return nil, BucketObjectLockConfigNotFound{Bucket: bucket}
	}
	configFile := path.Join(bucketConfigPrefix, bucket, objectLockConfig)
	configData, err := readConfig(ctx, objAPI, configFile)
	if err != nil {
		if err != errConfigNotFound {
			return nil, toObjectErr(err, bucket)
		}
		return objectlock.NewObjectLockConfig(), nil
	}
	cfg, err := objectlock.ParseObjectLockConfig(bytes.NewReader(configData))
	if err != nil {
		return nil, toObjectErr(err, bucket)
	}
	return cfg, nil
}

func saveBucketObjectLockConfig(ctx context.Context, objAPI ObjectLayer, bucket string, config *objectlock.Config) error {
	meta, err := loadBucketMetadata(ctx, objAPI, bucket)
	if err != nil && err != errMetaDataConverted {
		return toObjectErr(err, bucket)
	}
	if !meta.LockEnabled {
		return BucketObjectLockConfigNotFound{Bucket: bucket}
	}

	data, err := xml.Marshal(config)
	if err != nil {
		return toObjectErr(err, bucket)
	}

	configFile := path.Join(bucketConfigPrefix, bucket, objectLockConfig)
	if err = saveConfig(ctx, objAPI, configFile, data); err != nil {
		return toObjectErr(err, bucket)
	}
	return nil
}

// NewBucketObjectLockSys returns initialized BucketObjectLockSys
func NewBucketObjectLockSys() *BucketObjectLockSys {
	return &BucketObjectLockSys{
		retentionMap: make(map[string]*objectlock.Retention),
	}
}

// Init - initializes bucket object lock config system for all buckets
func (sys *BucketObjectLockSys) Init(buckets []BucketInfo, objAPI ObjectLayer) error {
	if objAPI == nil {
		return errServerNotInitialized
	}

	// In gateway mode, we always fetch the bucket object lock configuration from the gateway backend.
	// So, this is a no-op for gateway servers.
	if globalIsGateway {
		return nil
	}

	// Load BucketObjectLockSys once during boot.
	return sys.load(buckets, objAPI)
}

func (sys *BucketObjectLockSys) load(buckets []BucketInfo, objAPI ObjectLayer) error {
	for _, bucket := range buckets {
		ctx := logger.SetReqInfo(GlobalContext, &logger.ReqInfo{BucketName: bucket.Name})
		meta, err := loadBucketMetadata(ctx, objAPI, bucket.Name)
		if err != nil {
			if err != errMetaDataConverted {
				return err
			}

			meta.Created = bucket.Created
			logger.LogIf(ctx, meta.save(ctx, objAPI))
		}
		if !meta.LockEnabled {
			continue
		}

		configFile := path.Join(bucketConfigPrefix, bucket.Name, objectLockConfig)
		configData, err := readConfig(ctx, objAPI, configFile)
		if err != nil {
			if errors.Is(err, errConfigNotFound) {
				globalBucketObjectLockSys.Set(bucket.Name, &objectlock.Retention{})
				continue
			}
			return err
		}

		config, err := objectlock.ParseObjectLockConfig(bytes.NewReader(configData))
		if err != nil {
			return err
		}
		retention := &objectlock.Retention{}
		if config.Rule != nil {
			retention = config.ToRetention()
		}
		globalBucketObjectLockSys.Set(bucket.Name, retention)
	}
	return nil
}
