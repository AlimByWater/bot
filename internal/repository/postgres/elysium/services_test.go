package elysium_test

import (
	"context"
	"elysium/internal/entity"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateService(t *testing.T) {
	t.Skip()

	testCases := []struct {
		name        string
		service     entity.Service
		expectError bool
	}{
		{
			name: "Create new service",
			service: entity.Service{
				BotID:       1,
				Name:        "Test Service",
				Description: "Test Description",
				Price:       100,
				IsActive:    true,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createdService, err := elysiumRepo.CreateService(context.Background(), tc.service)
			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotZero(t, createdService.ID)

			t.Cleanup(func() {
				err := elysiumRepo.DeleteServiceHard(context.Background(), createdService.ID)
				require.NoError(t, err)
			})

			fetchedService, err := elysiumRepo.GetServiceByID(context.Background(), createdService.ID)
			require.NoError(t, err)
			require.Equal(t, tc.service.Name, fetchedService.Name)
			require.Equal(t, tc.service.Description, fetchedService.Description)
			require.Equal(t, tc.service.Price, fetchedService.Price)
			require.Equal(t, tc.service.IsActive, fetchedService.IsActive)
		})
	}
}

func TestUpdateService(t *testing.T) {
	t.Skip()

	// First create a service
	initialService := entity.Service{
		BotID:       1,
		Name:        "Initial Service",
		Description: "Initial Description",
		Price:       100,
		IsActive:    true,
	}

	createdService, err := elysiumRepo.CreateService(context.Background(), initialService)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := elysiumRepo.DeleteServiceHard(context.Background(), createdService.ID)
		require.NoError(t, err)
	})

	// Update the service
	updatedService := createdService
	updatedService.Name = "Updated Service"
	updatedService.Description = "Updated Description"
	updatedService.Price = 200
	updatedService.IsActive = false

	result, err := elysiumRepo.UpdateService(context.Background(), updatedService)
	require.NoError(t, err)
	require.NotNil(t, result.UpdatedAt)

	// Verify the update
	fetchedService, err := elysiumRepo.GetServiceByID(context.Background(), createdService.ID)
	require.NoError(t, err)
	require.Equal(t, updatedService.Name, fetchedService.Name)
	require.Equal(t, updatedService.Description, fetchedService.Description)
	require.Equal(t, updatedService.Price, fetchedService.Price)
	require.Equal(t, updatedService.IsActive, fetchedService.IsActive)
}

func TestGetServiceByID(t *testing.T) {
	t.Skip()

	// Create a service first
	service := entity.Service{
		BotID:       1,
		Name:        "Test Service",
		Description: "Test Description",
		Price:       100,
		IsActive:    true,
	}

	createdService, err := elysiumRepo.CreateService(context.Background(), service)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := elysiumRepo.DeleteServiceHard(context.Background(), createdService.ID)
		require.NoError(t, err)
	})

	// Test getting the service
	fetchedService, err := elysiumRepo.GetServiceByID(context.Background(), createdService.ID)
	require.NoError(t, err)
	require.Equal(t, createdService.ID, fetchedService.ID)
	require.Equal(t, service.Name, fetchedService.Name)
	require.Equal(t, service.Description, fetchedService.Description)
	require.Equal(t, service.Price, fetchedService.Price)
	require.Equal(t, service.IsActive, fetchedService.IsActive)

	// Test getting non-existent service
	_, err = elysiumRepo.GetServiceByID(context.Background(), 99999)
	require.Error(t, err)
}

func TestListServicesByBotID(t *testing.T) {
	t.Skip()

	botID := int64(1)

	// Create multiple services
	services := []entity.Service{
		{
			BotID:       botID,
			Name:        "Service 1",
			Description: "Description 1",
			Price:       100,
			IsActive:    true,
		},
		{
			BotID:       botID,
			Name:        "Service 2",
			Description: "Description 2",
			Price:       200,
			IsActive:    true,
		},
	}

	var createdIDs []int
	for _, service := range services {
		created, err := elysiumRepo.CreateService(context.Background(), service)
		require.NoError(t, err)
		createdIDs = append(createdIDs, created.ID)
	}

	t.Cleanup(func() {
		for _, id := range createdIDs {
			err := elysiumRepo.DeleteServiceHard(context.Background(), id)
			require.NoError(t, err)
		}
	})

	// Test listing services
	fetchedServices, err := elysiumRepo.ListServicesByBotID(context.Background(), botID)
	require.NoError(t, err)
	require.Len(t, fetchedServices, len(services))
}

func TestDeleteServiceHard(t *testing.T) {
	t.Skip()

	// Create a service first
	service := entity.Service{
		BotID:       1,
		Name:        "Test Service",
		Description: "Test Description",
		Price:       100,
		IsActive:    true,
	}

	createdService, err := elysiumRepo.CreateService(context.Background(), service)
	require.NoError(t, err)

	// Test deleting the service
	err = elysiumRepo.DeleteServiceHard(context.Background(), createdService.ID)
	require.NoError(t, err)

	// Verify the service is deleted
	_, err = elysiumRepo.GetServiceByID(context.Background(), createdService.ID)
	require.Error(t, err)

	// Test deleting non-existent service
	err = elysiumRepo.DeleteServiceHard(context.Background(), 99999)
	require.Error(t, err)
}
