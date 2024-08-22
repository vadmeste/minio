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
	"github.com/minio/pkg/v2/sync/errgroup"
)

const (
	peerS3Bucket            = "bucket"
	peerS3BucketDeleted     = "bucket-deleted"
	peerS3BucketForceCreate = "force-create"
	peerS3BucketForceDelete = "force-delete"
)

func healBucketLocal(ctx context.Context, bucket string, opts madmin.HealOpts) (res madmin.HealResultItem, err error) {
	globalLocalDrivesMu.RLock()
	localDrives := cloneDrives(globalLocalDrives)
	globalLocalDrivesMu.RUnlock()

	// Initialize sync waitgroup.
	g := errgroup.WithNErrs(len(localDrives))

	// Disk states slices
	beforeState := make([]string, len(localDrives))
	afterState := make([]string, len(localDrives))

	// Make a volume entry on all underlying storage disks.
	for index := range localDrives {
		index := index
		g.Go(func() (serr error) {
			if localDrives[index] == nil {
				beforeState[index] = madmin.DriveStateOffline
				afterState[index] = madmin.DriveStateOffline
				return errDiskNotFound
			}

			beforeState[index] = madmin.DriveStateOk
			afterState[index] = madmin.DriveStateOk

			if bucket == minioReservedBucket {
				return nil
			}

			_, serr = localDrives[index].StatVol(ctx, bucket)
			if serr != nil {
				if serr == errDiskNotFound {
					beforeState[index] = madmin.DriveStateOffline
					afterState[index] = madmin.DriveStateOffline
					return serr
				}
				if serr != errVolumeNotFound {
					beforeState[index] = madmin.DriveStateCorrupt
					afterState[index] = madmin.DriveStateCorrupt
					return serr
				}

				beforeState[index] = madmin.DriveStateMissing
				afterState[index] = madmin.DriveStateMissing

				return serr
			}
			return nil
		}, index)
	}

	errs := g.Wait()

	// Initialize heal result info
	res = madmin.HealResultItem{
		Type:      madmin.HealItemBucket,
		Bucket:    bucket,
		DiskCount: len(localDrives),
		SetCount:  -1, // explicitly set an invalid value -1, for bucket heal scenario
	}

	// mutate only if not a dry-run
	if opts.DryRun {
		return res, nil
	}

	for i := range beforeState {
		res.Before.Drives = append(res.Before.Drives, madmin.HealDriveInfo{
			UUID:     "",
			Endpoint: localDrives[i].String(),
			State:    beforeState[i],
		})
	}

	// check dangling and delete bucket only if its not a meta bucket
	if !isMinioMetaBucketName(bucket) && !isAllBucketsNotFound(errs) && opts.Remove {
		g := errgroup.WithNErrs(len(localDrives))
		for index := range localDrives {
			index := index
			g.Go(func() error {
				if localDrives[index] == nil {
					return errDiskNotFound
				}
				localDrives[index].DeleteVol(ctx, bucket, false)
				return nil
			}, index)
		}

		g.Wait()
	}

	// Create the quorum lost volume only if its nor makred for delete
	if !opts.Remove {
		// Initialize sync waitgroup.
		g = errgroup.WithNErrs(len(localDrives))

		// Make a volume entry on all underlying storage disks.
		for index := range localDrives {
			index := index
			g.Go(func() error {
				if beforeState[index] == madmin.DriveStateMissing {
					err := localDrives[index].MakeVol(ctx, bucket)
					if err == nil {
						afterState[index] = madmin.DriveStateOk
					}
					return err
				}
				return errs[index]
			}, index)
		}

		errs = g.Wait()
	}

	for i := range afterState {
		res.After.Drives = append(res.After.Drives, madmin.HealDriveInfo{
			UUID:     "",
			Endpoint: localDrives[i].String(),
			State:    afterState[i],
		})
	}
	return res, nil
}

func listBucketsLocal(ctx context.Context, opts BucketOptions) (buckets []BucketInfo, err error) {
	globalLocalDrivesMu.RLock()
	localDrives := cloneDrives(globalLocalDrives)
	globalLocalDrivesMu.RUnlock()

	quorum := (len(localDrives) / 2)

	buckets = make([]BucketInfo, 0, 32)
	healBuckets := map[string]VolInfo{}

	// lists all unique buckets across drives.
	if err := listAllBuckets(ctx, localDrives, healBuckets, quorum); err != nil {
		return nil, err
	}

	// include deleted buckets in listBuckets output
	deletedBuckets := map[string]VolInfo{}

	if opts.Deleted {
		// lists all deleted buckets across drives.
		if err := listDeletedBuckets(ctx, localDrives, deletedBuckets, quorum); err != nil {
			return nil, err
		}
	}

	for _, v := range healBuckets {
		bi := BucketInfo{
			Name:    v.Name,
			Created: v.Created,
		}
		if vi, ok := deletedBuckets[v.Name]; ok {
			bi.Deleted = vi.Created
		}
		buckets = append(buckets, bi)
	}

	for _, v := range deletedBuckets {
		if _, ok := healBuckets[v.Name]; !ok {
			buckets = append(buckets, BucketInfo{
				Name:    v.Name,
				Deleted: v.Created,
			})
		}
	}

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

	return reduceWriteQuorumErrs(ctx, g.Wait(), bucketOpIgnoredErrs, (len(localDrives)/2)+1)
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
