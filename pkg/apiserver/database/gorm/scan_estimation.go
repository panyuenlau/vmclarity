// Copyright © 2023 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gorm

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/apiserver/common"
	"github.com/openclarity/vmclarity/pkg/apiserver/database/types"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

const (
	scanEstimationSchemaName = "ScanEstimation"
)

type ScanEstimation struct {
	ODataObject
}

type ScanEstimationsTableHandler struct {
	DB *gorm.DB
}

func (db *Handler) ScanEstimationsTable() types.ScanEstimationsTable {
	return &ScanEstimationsTableHandler{
		DB: db.DB,
	}
}

func (s *ScanEstimationsTableHandler) GetScanEstimations(params models.GetScanEstimationsParams) (models.ScanEstimations, error) {
	var scanEstimations []ScanEstimation
	err := ODataQuery(s.DB, scanEstimationSchemaName, params.Filter, params.Select, params.Expand, params.OrderBy, params.Top, params.Skip, true, &scanEstimations)
	if err != nil {
		return models.ScanEstimations{}, err
	}

	items := make([]models.ScanEstimation, len(scanEstimations))
	for i, sc := range scanEstimations {
		var scanEstimation models.ScanEstimation
		err = json.Unmarshal(sc.Data, &scanEstimation)
		if err != nil {
			return models.ScanEstimations{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		items[i] = scanEstimation
	}

	output := models.ScanEstimations{Items: &items}

	if params.Count != nil && *params.Count {
		count, err := ODataCount(s.DB, scanEstimationSchemaName, params.Filter)
		if err != nil {
			return models.ScanEstimations{}, fmt.Errorf("failed to count records: %w", err)
		}
		output.Count = &count
	}

	return output, nil
}

func (s *ScanEstimationsTableHandler) GetScanEstimation(scanEstimationID models.ScanEstimationID, params models.GetScanEstimationsScanEstimationIDParams) (models.ScanEstimation, error) {
	var dbScanEstimation ScanEstimation
	filter := fmt.Sprintf("id eq '%s'", scanEstimationID)
	err := ODataQuery(s.DB, scanEstimationSchemaName, &filter, params.Select, params.Expand, nil, nil, nil, false, &dbScanEstimation)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.ScanEstimation{}, types.ErrNotFound
		}
		return models.ScanEstimation{}, err
	}

	var apiScanEstimation models.ScanEstimation
	err = json.Unmarshal(dbScanEstimation.Data, &apiScanEstimation)
	if err != nil {
		return models.ScanEstimation{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiScanEstimation, nil
}

func (s *ScanEstimationsTableHandler) CreateScanEstimation(scanEstimation models.ScanEstimation) (models.ScanEstimation, error) {
	// Check the user didn't provide an ID
	if scanEstimation.Id != nil {
		return models.ScanEstimation{}, &common.BadRequestError{
			Reason: "can not specify id field when creating a new ScanEstimation",
		}
	}

	// Generate a new UUID
	scanEstimation.Id = utils.PointerTo(uuid.New().String())

	// Initialise revision
	scanEstimation.Revision = utils.PointerTo(1)

	marshaled, err := json.Marshal(scanEstimation)
	if err != nil {
		return models.ScanEstimation{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	newScanEstimation := ScanEstimation{}
	newScanEstimation.Data = marshaled

	if err = s.DB.Create(&newScanEstimation).Error; err != nil {
		return models.ScanEstimation{}, fmt.Errorf("failed to create scan estimation in db: %w", err)
	}

	var apiScanEstimation models.ScanEstimation
	err = json.Unmarshal(newScanEstimation.Data, &apiScanEstimation)
	if err != nil {
		return models.ScanEstimation{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiScanEstimation, nil
}

// nolint:cyclop
func (s *ScanEstimationsTableHandler) SaveScanEstimation(scanEstimation models.ScanEstimation, params models.PutScanEstimationsScanEstimationIDParams) (models.ScanEstimation, error) {
	if scanEstimation.Id == nil || *scanEstimation.Id == "" {
		return models.ScanEstimation{}, &common.BadRequestError{
			Reason: "id is required to save scan estimation",
		}
	}

	var dbObj ScanEstimation
	if err := getExistingObjByID(s.DB, scanEstimationSchemaName, *scanEstimation.Id, &dbObj); err != nil {
		return models.ScanEstimation{}, fmt.Errorf("failed to get scan estimation from db: %w", err)
	}

	var dbScanEstimation models.ScanEstimation
	if err := json.Unmarshal(dbObj.Data, &dbScanEstimation); err != nil {
		return models.ScanEstimation{}, fmt.Errorf("failed to convert DB object to API model: %w", err)
	}

	if err := checkRevisionEtag(params.IfMatch, dbScanEstimation.Revision); err != nil {
		return models.ScanEstimation{}, err
	}

	scanEstimation.Revision = bumpRevision(dbScanEstimation.Revision)

	marshaled, err := json.Marshal(scanEstimation)
	if err != nil {
		return models.ScanEstimation{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	dbObj.Data = marshaled

	if err = s.DB.Save(&dbObj).Error; err != nil {
		return models.ScanEstimation{}, fmt.Errorf("failed to save scan estimation in db: %w", err)
	}

	var apiScanEstimation models.ScanEstimation
	if err = json.Unmarshal(dbObj.Data, &apiScanEstimation); err != nil {
		return models.ScanEstimation{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiScanEstimation, nil
}

// nolint:cyclop
func (s *ScanEstimationsTableHandler) UpdateScanEstimation(scanEstimation models.ScanEstimation, params models.PatchScanEstimationsScanEstimationIDParams) (models.ScanEstimation, error) {
	if scanEstimation.Id == nil || *scanEstimation.Id == "" {
		return models.ScanEstimation{}, &common.BadRequestError{
			Reason: "id is required to update scan estimation",
		}
	}

	var dbObj ScanEstimation
	if err := getExistingObjByID(s.DB, scanEstimationSchemaName, *scanEstimation.Id, &dbObj); err != nil {
		return models.ScanEstimation{}, err
	}

	var dbScanEstimation models.ScanEstimation
	if err := json.Unmarshal(dbObj.Data, &dbScanEstimation); err != nil {
		return models.ScanEstimation{}, fmt.Errorf("failed to convert DB object to API model: %w", err)
	}

	if err := checkRevisionEtag(params.IfMatch, dbScanEstimation.Revision); err != nil {
		return models.ScanEstimation{}, err
	}

	scanEstimation.Revision = bumpRevision(dbScanEstimation.Revision)

	var err error
	dbObj.Data, err = patchObject(dbObj.Data, scanEstimation)
	if err != nil {
		return models.ScanEstimation{}, fmt.Errorf("failed to apply patch: %w", err)
	}

	var ret models.ScanEstimation
	err = json.Unmarshal(dbObj.Data, &ret)
	if err != nil {
		return models.ScanEstimation{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	if err := s.DB.Save(&dbObj).Error; err != nil {
		return models.ScanEstimation{}, fmt.Errorf("failed to save scan estimation in db: %w", err)
	}

	return ret, nil
}

func (s *ScanEstimationsTableHandler) DeleteScanEstimation(scanEstimationID models.ScanEstimationID) error {
	if err := deleteObjByID(s.DB, scanEstimationID, &ScanEstimation{}); err != nil {
		return fmt.Errorf("failed to delete scan estimation: %w", err)
	}

	return nil
}
