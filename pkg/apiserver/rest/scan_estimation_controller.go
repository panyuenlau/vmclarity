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

package rest

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/apiserver/common"
	databaseTypes "github.com/openclarity/vmclarity/pkg/apiserver/database/types"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

func (s *ServerImpl) GetScanEstimations(ctx echo.Context, params models.GetScanEstimationsParams) error {
	scanEstimations, err := s.dbHandler.ScanEstimationsTable().GetScanEstimations(params)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan estimations from db: %v", err))
	}

	return sendResponse(ctx, http.StatusOK, scanEstimations)
}

func (s *ServerImpl) PostScanEstimations(ctx echo.Context) error {
	var scanEstimation models.ScanEstimation
	err := ctx.Bind(&scanEstimation)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	createdScanEstimation, err := s.dbHandler.ScanEstimationsTable().CreateScanEstimation(scanEstimation)
	if err != nil {
		var validationErr *common.BadRequestError
		switch {
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to create scan estimation in db: %v", err))
		}
	}

	return sendResponse(ctx, http.StatusCreated, createdScanEstimation)
}

func (s *ServerImpl) DeleteScanEstimationsScanEstimationID(ctx echo.Context, scanEstimationID models.ScanEstimationID) error {
	success := models.Success{
		Message: utils.PointerTo(fmt.Sprintf("scan estimation %v deleted", scanEstimationID)),
	}

	if err := s.dbHandler.ScanEstimationsTable().DeleteScanEstimation(scanEstimationID); err != nil {
		if errors.Is(err, databaseTypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("ScanEstimation with ID %v not found", scanEstimationID))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to delete scan estimation from db. scanEstimationID=%v: %v", scanEstimationID, err))
	}

	return sendResponse(ctx, http.StatusOK, &success)
}

func (s *ServerImpl) GetScanEstimationsScanEstimationID(ctx echo.Context, scanEstimationID models.ScanEstimationID, params models.GetScanEstimationsScanEstimationIDParams) error {
	scanEstimation, err := s.dbHandler.ScanEstimationsTable().GetScanEstimation(scanEstimationID, params)
	if err != nil {
		if errors.Is(err, databaseTypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("ScanEstimation with ID %v not found", scanEstimationID))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan estimation from db. id=%v: %v", scanEstimationID, err))
	}
	return sendResponse(ctx, http.StatusOK, scanEstimation)
}

func (s *ServerImpl) PatchScanEstimationsScanEstimationID(ctx echo.Context, scanEstimationID models.ScanEstimationID, params models.PatchScanEstimationsScanEstimationIDParams) error {
	var scanEstimation models.ScanEstimation
	err := ctx.Bind(&scanEstimation)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PATCH request might not contain the ID in the body, so set it from
	// the URL field so that the DB layer knows which object is being updated.
	if scanEstimation.Id != nil && *scanEstimation.Id != scanEstimationID {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("id in body %s does not match object %s to be updated", *scanEstimation.Id, scanEstimationID))
	}
	scanEstimation.Id = &scanEstimationID

	updatedScanEstimation, err := s.dbHandler.ScanEstimationsTable().UpdateScanEstimation(scanEstimation, params)
	if err != nil {
		var validationErr *common.BadRequestError
		var preconditionFailedErr *databaseTypes.PreconditionFailedError
		switch {
		case errors.Is(err, databaseTypes.ErrNotFound):
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("ScanEstimation with ID %v not found", scanEstimationID))
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		case errors.As(err, &preconditionFailedErr):
			return sendError(ctx, http.StatusPreconditionFailed, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update scan estimation in db. scanEstimationID=%v: %v", scanEstimationID, err))
		}
	}

	return sendResponse(ctx, http.StatusOK, updatedScanEstimation)
}

func (s *ServerImpl) PutScanEstimationsScanEstimationID(ctx echo.Context, scanEstimationID models.ScanEstimationID, params models.PutScanEstimationsScanEstimationIDParams) error {
	var scanEstimation models.ScanEstimation
	err := ctx.Bind(&scanEstimation)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PUT request might not contain the ID in the body, so set it from the
	// URL field so that the DB layer knows which object is being updated.
	if scanEstimation.Id != nil && *scanEstimation.Id != scanEstimationID {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("id in body %s does not match object %s to be updated", *scanEstimation.Id, scanEstimationID))
	}
	scanEstimation.Id = &scanEstimationID

	updatedScanEstimation, err := s.dbHandler.ScanEstimationsTable().SaveScanEstimation(scanEstimation, params)
	if err != nil {
		var validationErr *common.BadRequestError
		var preconditionFailedErr *databaseTypes.PreconditionFailedError
		switch {
		case errors.Is(err, databaseTypes.ErrNotFound):
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("ScanEstimation with ID %v not found", scanEstimationID))
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		case errors.As(err, &preconditionFailedErr):
			return sendError(ctx, http.StatusPreconditionFailed, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to save scan estimation in db. scanEstimationID=%v: %v", scanEstimationID, err))
		}
	}

	return sendResponse(ctx, http.StatusOK, updatedScanEstimation)
}
