// Copyright (C) 2019-2021 Zilliz. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under the License.

package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/milvus-io/milvus-sdk-go/v2/internal/proto/common"
	"github.com/milvus-io/milvus-sdk-go/v2/internal/proto/server"
)

// CreatePartition create partition for collection
func (c *grpcClient) CreatePartition(ctx context.Context, collName string, partitionName string) error {
	if c.service == nil {
		return ErrClientNotReady
	}
	if err := c.checkCollectionExists(ctx, collName); err != nil {
		return err
	}
	has, err := c.HasPartition(ctx, collName, partitionName)
	if err != nil {
		return err
	}
	if has {
		return fmt.Errorf("partition %s of collection %s already exists", partitionName, collName)
	}

	req := &server.CreatePartitionRequest{
		DbName:         "", //reserved
		CollectionName: collName,
		PartitionName:  partitionName,
	}
	resp, err := c.service.CreatePartition(ctx, req)
	if err != nil {
		return err
	}
	return handleRespStatus(resp)
}

func (c *grpcClient) checkPartitionExists(ctx context.Context, collName string, partitionName string) error {
	err := c.checkCollectionExists(ctx, collName)
	if err != nil {
		return err
	}

	has, err := c.HasPartition(ctx, collName, partitionName)
	if err != nil {
		return err
	}
	if !has {
		return partNotExistsErr(collName, partitionName)
	}
	return nil
}

// DropPartition drop partition from collection
func (c *grpcClient) DropPartition(ctx context.Context, collName string, partitionName string) error {
	if c.service == nil {
		return ErrClientNotReady
	}
	if err := c.checkPartitionExists(ctx, collName, partitionName); err != nil {
		return err
	}
	req := &server.DropPartitionRequest{
		DbName:         "",
		CollectionName: collName,
		PartitionName:  partitionName,
	}
	resp, err := c.service.DropPartition(ctx, req)
	if err != nil {
		return err
	}
	return handleRespStatus(resp)
}

// HasPartition check whether specified partition exists
func (c *grpcClient) HasPartition(ctx context.Context, collName string, partitionName string) (bool, error) {
	if c.service == nil {
		return false, ErrClientNotReady
	}
	req := &server.HasPartitionRequest{
		DbName:         "", // reserved
		CollectionName: collName,
		PartitionName:  partitionName,
	}
	resp, err := c.service.HasPartition(ctx, req)
	if err != nil {
		return false, err
	}
	if resp.GetStatus().GetErrorCode() != common.ErrorCode_Success {
		return false, errors.New("request failed")
	}
	return resp.GetValue(), nil
}

// ShowPartitions list all paritions from collection
func (c *grpcClient) ShowPartitions(ctx context.Context, collName string) ([]*entity.Partition, error) {
	if c.service == nil {
		return []*entity.Partition{}, ErrClientNotReady
	}
	req := &server.ShowPartitionsRequest{
		DbName:         "", // reserved
		CollectionName: collName,
	}
	resp, err := c.service.ShowPartitions(ctx, req)
	if err != nil {
		return []*entity.Partition{}, err
	}
	partitions := make([]*entity.Partition, 0, len(resp.GetPartitionIDs()))
	for idx, partitionID := range resp.GetPartitionIDs() {
		partitions = append(partitions, &entity.Partition{ID: partitionID, Name: resp.GetPartitionNames()[idx]})
	}
	return partitions, nil
}

// LoadPartitions load collection paritions into memory
func (c *grpcClient) LoadPartitions(ctx context.Context, collName string, partitionNames []string, async bool) error {
	if c.service == nil {
		return ErrClientNotReady
	}
	for _, partitionName := range partitionNames {
		if err := c.checkPartitionExists(ctx, collName, partitionName); err != nil {
			return err
		}
	}
	partitions, err := c.ShowPartitions(ctx, collName)
	if err != nil {
		return err
	}
	m := make(map[string]int64)
	for _, partition := range partitions {
		m[partition.Name] = partition.ID
	}
	// load partitions ids
	ids := make(map[int64]struct{})
	for _, partitionName := range partitionNames {
		id, has := m[partitionName]
		if !has {
			return fmt.Errorf("Collection %s does not has partitions %s", collName, partitionName)
		}
		ids[id] = struct{}{}
	}

	req := &server.LoadPartitionsRequest{
		DbName:         "", // reserved
		CollectionName: collName,
		PartitionNames: partitionNames,
	}
	resp, err := c.service.LoadPartitions(ctx, req)
	if err != nil {
		return err
	}
	if err := handleRespStatus(resp); err != nil {
		return err
	}

	if !async {
		segments, _ := c.GetPersistentSegmentInfo(ctx, collName)
		target := make(map[int64]*entity.Segment)
		for _, segment := range segments {
			if segment.NumRows == 0 {
				continue
			}
			if _, has := ids[segment.ParititionID]; !has {
				// segment not belongs to partition
				continue
			}
			target[segment.ID] = segment
		}
		for len(target) > 0 {
			current, err := c.GetQuerySegmentInfo(ctx, collName)
			if err == nil {
				for _, segment := range current {
					ts, has := target[segment.ID]
					if has {
						if segment.NumRows >= ts.NumRows {
							delete(target, segment.ID)
						}
					}
				}
			}
			time.Sleep(time.Millisecond * 100)
		}
	}

	return nil
}

// ReleasePartitions release partitions
func (c *grpcClient) ReleasePartitions(ctx context.Context, collName string, partitionNames []string) error {
	if c.service == nil {
		return ErrClientNotReady
	}
	for _, partitionName := range partitionNames {
		if err := c.checkPartitionExists(ctx, collName, partitionName); err != nil {
			return err
		}
	}
	req := &server.ReleasePartitionsRequest{
		DbName:         "", // reserved
		CollectionName: collName,
		PartitionNames: partitionNames,
	}
	resp, err := c.service.ReleasePartitions(ctx, req)
	if err != nil {
		return err
	}
	if err := handleRespStatus(resp); err != nil {
		return err
	}

	return nil
}
