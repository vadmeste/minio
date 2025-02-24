// Copyright (c) 2015-2022 MinIO, Inc.
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
	"context"
	"errors"

	"github.com/minio/madmin-go/v3"
	"github.com/minio/pkg/v3/sync/errgroup"
	"github.com/puzpuzpuz/xsync/v3"
)

const (
	peerS3Bucket            = "bucket"
	peerS3BucketDeleted     = "bucket-deleted"
	peerS3BucketStale       = "bucket-stale"
	peerS3BucketForceCreate = "force-create"
	peerS3BucketForceDelete = "force-delete"
)

func healBucketLocal(ctx context.Context, bucket string, stale bool, opts madmin.HealOpts) (res madmin.HealResultItem, err error) {
	globalLocalDrivesMu.RLock()
	localDrives := cloneDrives(globalLocalDrives)
	globalLocalDrivesMu.RUnlock()

	// Initialize heal result info
	res = madmin.HealResultItem{
		Type:      madmin.HealItemBucket,
		Bucket:    bucket,
		DiskCount: len(localDrives),
		SetCount:  -1, // explicitly set an invalid value -1, for bucket heal scenario
	}

	res.Before.Drives = make([]madmin.HealDriveInfo, len(localDrives))
	res.After.Drives = make([]madmin.HealDriveInfo, len(localDrives))

	for i := range localDrives {
		if localDrives[i] != nil {
			res.Before.Drives[i].Endpoint = localDrives[i].String()
			res.After.Drives[i].Endpoint = localDrives[i].String()
		}
	}

	// Initialize sync waitgroup.
	g := errgroup.WithNErrs(len(localDrives))

	// Make a volume entry on all underlying storage disks.
	for index := range localDrives {
		index := index
		g.Go(func() error {
			var state string
			defer func() {
				res.Before.Drives[index].State = state
				res.After.Drives[index].State = state
			}()
			if localDrives[index] == nil {
				state = madmin.DriveStateOffline
				return errDiskNotFound
			}
			_, err := localDrives[index].StatVol(ctx, bucket)
			switch {
			case err == nil:
				state = madmin.DriveStateOk
			case errors.Is(err, errDiskNotFound):
				state = madmin.DriveStateOffline
			case errors.Is(err, errVolumeNotFound):
				state = madmin.DriveStateMissing
			default:
				state = madmin.DriveStateUnknown
			}
			return err
		}, index)
	}

	g.Wait()

	// Make the after state same as the before state if mutation are not allowed
	if opts.DryRun || stale && (!opts.Remove || isMinioMetaBucketName(bucket)) {
		return res, nil
	}

	g = errgroup.WithNErrs(len(localDrives))
	if stale {
		for index := range localDrives {
			index := index
			g.Go(func() error {
				if localDrives[index] == nil {
					return errDiskNotFound
				}
				return localDrives[index].DeleteVol(ctx, bucket, false)
			}, index)
		}
	} else {
		// Make a volume entry on all underlying storage disks.
		for index := range localDrives {
			index := index
			g.Go(func() (err error) {
				if localDrives[index] == nil {
					return errDiskNotFound
				}
				return localDrives[index].MakeVol(ctx, bucket)
			}, index)
		}
	}

	errs := g.Wait()
	for i, e := range errs {
		switch {
		case errors.Is(e, errDiskNotFound):
			res.After.Drives[i].State = madmin.DriveStateOffline
		case e == nil:
			fallthrough
		case errors.Is(e, errVolumeNotFound):
			fallthrough
		case errors.Is(e, errVolumeNotEmpty):
			fallthrough
		case errors.Is(e, errVolumeExists):
			res.After.Drives[i].State = madmin.DriveStateOk
		default:
			res.After.Drives[i].State = madmin.DriveStateUnknown
		}
	}

	return res, nil
}

func listBucketsLocal(ctx context.Context, opts BucketOptions) (buckets []BucketInfo, err error) {
	globalLocalDrivesMu.RLock()
	localDrives := cloneDrives(globalLocalDrives)
	globalLocalDrivesMu.RUnlock()

	quorum := (len(localDrives) / 2)

	buckets = make([]BucketInfo, 0, 32)
	healBuckets := xsync.NewMapOf[string, VolInfo]()

	// lists all unique buckets across drives.
	if err := listAllBuckets(ctx, localDrives, healBuckets, quorum); err != nil {
		return nil, err
	}

	// include deleted buckets in listBuckets output
	deletedBuckets := xsync.NewMapOf[string, VolInfo]()

	if opts.Deleted {
		// lists all deleted buckets across drives.
		if err := listDeletedBuckets(ctx, localDrives, deletedBuckets, quorum); err != nil {
			return nil, err
		}
	}

	healBuckets.Range(func(_ string, volInfo VolInfo) bool {
		bi := BucketInfo{
			Name:    volInfo.Name,
			Created: volInfo.Created,
		}
		if vi, ok := deletedBuckets.Load(volInfo.Name); ok {
			bi.Deleted = vi.Created
		}
		buckets = append(buckets, bi)
		return true
	})

	deletedBuckets.Range(func(_ string, v VolInfo) bool {
		if _, ok := healBuckets.Load(v.Name); !ok {
			buckets = append(buckets, BucketInfo{
				Name:    v.Name,
				Deleted: v.Created,
			})
		}
		return true
	})

	return buckets, nil
}

func cloneDrives(drives []StorageAPI) []StorageAPI {
	newDrives := make([]StorageAPI, len(drives))
	copy(newDrives, drives)
	return newDrives
}

func getBucketInfoLocal(ctx context.Context, bucket string, opts BucketOptions) (BucketInfo, error) {
	globalLocalDrivesMu.RLock()
	localDrives := cloneDrives(globalLocalDrives)
	globalLocalDrivesMu.RUnlock()

	g := errgroup.WithNErrs(len(localDrives)).WithConcurrency(32)
	bucketsInfo := make([]BucketInfo, len(localDrives))

	// Make a volume entry on all underlying storage disks.
	for index := range localDrives {
		index := index
		g.Go(func() error {
			if localDrives[index] == nil {
				return errDiskNotFound
			}
			volInfo, err := localDrives[index].StatVol(ctx, bucket)
			if err != nil {
				if opts.Deleted {
					dvi, derr := localDrives[index].StatVol(ctx, pathJoin(minioMetaBucket, bucketMetaPrefix, deletedBucketsPrefix, bucket))
					if derr != nil {
						return err
					}
					bucketsInfo[index] = BucketInfo{Name: bucket, Deleted: dvi.Created}
					return nil
				}
				return err
			}

			bucketsInfo[index] = BucketInfo{Name: bucket, Created: volInfo.Created}
			return nil
		}, index)
	}

	errs := g.Wait()
	if err := reduceReadQuorumErrs(ctx, errs, bucketOpIgnoredErrs, (len(localDrives) / 2)); err != nil {
		return BucketInfo{}, err
	}

	var bucketInfo BucketInfo
	for i, err := range errs {
		if err == nil {
			bucketInfo = bucketsInfo[i]
			break
		}
	}

	return bucketInfo, nil
}

func deleteBucketLocal(ctx context.Context, bucket string, opts DeleteBucketOptions) error {
	globalLocalDrivesMu.RLock()
	localDrives := cloneDrives(globalLocalDrives)
	globalLocalDrivesMu.RUnlock()

	g := errgroup.WithNErrs(len(localDrives)).WithConcurrency(32)

	// Make a volume entry on all underlying storage disks.
	for index := range localDrives {
		index := index
		g.Go(func() error {
			if localDrives[index] == nil {
				return errDiskNotFound
			}
			return localDrives[index].DeleteVol(ctx, bucket, opts.Force)
		}, index)
	}

	var recreate bool
	errs := g.Wait()
	for index, err := range errs {
		if errors.Is(err, errVolumeNotEmpty) {
			recreate = true
		}
		if err == nil && recreate {
			// ignore any errors
			localDrives[index].MakeVol(ctx, bucket)
		}
	}

	// Since we recreated buckets and error was `not-empty`, return not-empty.
	if recreate {
		return errVolumeNotEmpty
	} // for all other errors reduce by write quorum.

	return reduceWriteQuorumErrs(ctx, errs, bucketOpIgnoredErrs, (len(localDrives)/2)+1)
}

func makeBucketLocal(ctx context.Context, bucket string, opts MakeBucketOptions) error {
	globalLocalDrivesMu.RLock()
	localDrives := cloneDrives(globalLocalDrives)
	globalLocalDrivesMu.RUnlock()

	g := errgroup.WithNErrs(len(localDrives)).WithConcurrency(32)

	// Make a volume entry on all underlying storage disks.
	for index := range localDrives {
		index := index
		g.Go(func() error {
			if localDrives[index] == nil {
				return errDiskNotFound
			}
			err := localDrives[index].MakeVol(ctx, bucket)
			if opts.ForceCreate && errors.Is(err, errVolumeExists) {
				// No need to return error when force create was
				// requested.
				return nil
			}
			return err
		}, index)
	}

	errs := g.Wait()
	return reduceWriteQuorumErrs(ctx, errs, bucketOpIgnoredErrs, (len(localDrives)/2)+1)
}
