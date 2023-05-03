//go:build !integration
// +build !integration

package app

import (
	"errors"
	"testing"
	"time"

	storage "github.com/skolzkyi/hwOTUS_YIA/hw12_13_14_15_calendar/internal/storage/event"
	"github.com/stretchr/testify/require"
)

func createTestEvent() *storage.Event {
	std := time.Date(2023, 3, 27, 0, 0, 0, 1, time.Local)
	emtd := 4 * time.Hour
	tEvent := storage.Event{
		ID:                    0,
		Title:                 "TestTitle",
		UserID:                "USER0",
		Description:           "abc",
		DateStart:             std,
		DateStop:              std.Add(7 * time.Hour),
		EventMessageTimeDelta: emtd,
	}
	return &tEvent
}

func TestSimpleValidators(t *testing.T) {
	t.Run("PositiveValidator", func(t *testing.T) {
		t.Parallel()
		tEvent := createTestEvent()
		_, err := SimpleEventValidator(tEvent.Title, tEvent.UserID, tEvent.Description, tEvent.DateStart, tEvent.DateStop, tEvent.EventMessageTimeDelta) //nolint:lll
		require.NoError(t, err)
	})
	t.Run("NegativeErrVoidTitle", func(t *testing.T) {
		t.Parallel()
		tEvent := createTestEvent()
		tEvent.Title = ""
		_, err := SimpleEventValidator(tEvent.Title, tEvent.UserID, tEvent.Description, tEvent.DateStart, tEvent.DateStop, tEvent.EventMessageTimeDelta) //nolint:lll
		require.Truef(t, errors.Is(err, ErrVoidTitle), "actual error %q", err)
	})
	t.Run("NegativeErrVoidUserID", func(t *testing.T) {
		t.Parallel()
		tEvent := createTestEvent()
		tEvent.UserID = ""
		_, err := SimpleEventValidator(tEvent.Title, tEvent.UserID, tEvent.Description, tEvent.DateStart, tEvent.DateStop, tEvent.EventMessageTimeDelta) //nolint:lll
		require.Truef(t, errors.Is(err, ErrVoidUserID), "actual error %q", err)
	})
	t.Run("NegativeErrVoidDateStart", func(t *testing.T) {
		t.Parallel()
		tEvent := createTestEvent()
		var voidTime time.Time
		tEvent.DateStart = voidTime
		_, err := SimpleEventValidator(tEvent.Title, tEvent.UserID, tEvent.Description, tEvent.DateStart, tEvent.DateStop, tEvent.EventMessageTimeDelta) //nolint:lll
		require.Truef(t, errors.Is(err, ErrVoidDateStart), "actual error %q", err)
	})
	t.Run("NegativErrVoidDateStop", func(t *testing.T) {
		t.Parallel()
		tEvent := createTestEvent()
		var voidTime time.Time
		tEvent.DateStop = voidTime
		_, err := SimpleEventValidator(tEvent.Title, tEvent.UserID, tEvent.Description, tEvent.DateStart, tEvent.DateStop, tEvent.EventMessageTimeDelta) //nolint:lll
		require.Truef(t, errors.Is(err, ErrVoidDateStop), "actual error %q", err)
	})
	t.Run("NegativErrTitleTooLong", func(t *testing.T) {
		t.Parallel()
		tEvent := createTestEvent()
		voidBuf := make([]byte, 255)
		tEvent.Title = string(voidBuf)
		_, err := SimpleEventValidator(tEvent.Title, tEvent.UserID, tEvent.Description, tEvent.DateStart, tEvent.DateStop, tEvent.EventMessageTimeDelta) //nolint:lll
		require.Truef(t, errors.Is(err, ErrTitleTooLong), "actual error %q", err)
	})
	t.Run("NegativErrUserIDTooLong", func(t *testing.T) {
		t.Parallel()
		tEvent := createTestEvent()
		voidBuf := make([]byte, 50)
		tEvent.UserID = string(voidBuf)
		_, err := SimpleEventValidator(tEvent.Title, tEvent.UserID, tEvent.Description, tEvent.DateStart, tEvent.DateStop, tEvent.EventMessageTimeDelta) //nolint:lll
		require.Truef(t, errors.Is(err, ErrUserIDTooLong), "actual error %q", err)
	})
	t.Run("NegativErrDescTooLong", func(t *testing.T) {
		t.Parallel()
		tEvent := createTestEvent()
		voidBuf := make([]byte, 1500)
		tEvent.Description = string(voidBuf)
		_, err := SimpleEventValidator(tEvent.Title, tEvent.UserID, tEvent.Description, tEvent.DateStart, tEvent.DateStop, tEvent.EventMessageTimeDelta) //nolint:lll
		require.Truef(t, errors.Is(err, ErrDescTooLong), "actual error %q", err)
	})
	t.Run("NegativErrEndDateBefstartDate", func(t *testing.T) {
		t.Parallel()
		tEvent := createTestEvent()
		tempTime := tEvent.DateStart //nolint:gocritic
		tEvent.DateStart = tEvent.DateStop
		tEvent.DateStop = tempTime
		_, err := SimpleEventValidator(tEvent.Title, tEvent.UserID, tEvent.Description, tEvent.DateStart, tEvent.DateStop, tEvent.EventMessageTimeDelta) //nolint:lll
		require.Truef(t, errors.Is(err, ErrEndDateBefstartDate), "actual error %q", err)
	})
}
