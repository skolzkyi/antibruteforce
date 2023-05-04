//go:build !integration
// +build !integration

package app

import (
	"context"
	"errors"
	"testing"
	"time"

	logger "github.com/skolzkyi/hwOTUS_YIA/hw12_13_14_15_calendar/internal/logger"
	storage "github.com/skolzkyi/hwOTUS_YIA/hw12_13_14_15_calendar/internal/storage/event"
	"github.com/stretchr/testify/require"
)

const messageTestUpdated = "testUpdated"

func initStorageInMemory(t *testing.T) *Storage {
	t.Helper()
	logger, _ := logger.New("debug")
	storage := New()
	err := storage.Init(context.Background(), logger, nil)
	require.NoError(t, err)
	return storage
}

func createTestEventPack(t *testing.T, s *Storage) {
	t.Helper()
	logger, _ := logger.New("debug")
	events := make([]storage.Event, 10)
	ctx := context.Background()
	// std := helpers.DateStartTime(time.Now())
	std := time.Date(2023, 3, 27, 0, 0, 0, 1, time.Local)
	emtd := 4 * time.Hour
	events[0] = storage.Event{
		ID:                    0,
		Title:                 "test0 - baseDateTime",
		UserID:                "USER0",
		Description:           "",
		DateStart:             std.Add(5 * time.Hour),
		DateStop:              std.Add(7 * time.Hour),
		EventMessageTimeDelta: emtd,
	}
	events[1] = storage.Event{
		ID:                    0,
		Title:                 "test1 - in this day",
		UserID:                "USER0",
		Description:           "",
		DateStart:             std.Add(9 * time.Hour),
		DateStop:              std.Add(11 * time.Hour),
		EventMessageTimeDelta: emtd,
	}
	events[2] = storage.Event{
		ID:                    0,
		Title:                 "test2 - in this week",
		UserID:                "USER0",
		Description:           "",
		DateStart:             std.Add(27 * time.Hour),
		DateStop:              std.Add(28 * time.Hour),
		EventMessageTimeDelta: emtd,
	}
	events[3] = storage.Event{
		ID:                    0,
		Title:                 "test3 - in this month",
		UserID:                "USER0",
		Description:           "",
		DateStart:             std.Add(192 * time.Hour),
		DateStop:              std.Add(193 * time.Hour),
		EventMessageTimeDelta: emtd,
	}
	events[4] = storage.Event{
		ID:                    0,
		Title:                 "test4 - on  in border of this day",
		UserID:                "USER0",
		Description:           "",
		DateStart:             std.Add(-3 * time.Hour),
		DateStop:              std.Add(1 * time.Hour),
		EventMessageTimeDelta: emtd,
	}
	events[5] = storage.Event{
		ID:                    0,
		Title:                 "test5 - on  out border of this day",
		UserID:                "USER0",
		Description:           "",
		DateStart:             std.Add(23 * time.Hour),
		DateStop:              std.Add(25 * time.Hour),
		EventMessageTimeDelta: emtd,
	}

	events[6] = storage.Event{
		ID:                    0,
		Title:                 "test6 - on  out border of this week",
		UserID:                "USER0",
		Description:           "",
		DateStart:             std.Add(167 * time.Hour),
		DateStop:              std.Add(170 * time.Hour),
		EventMessageTimeDelta: emtd,
	}
	events[7] = storage.Event{
		ID:                    0,
		Title:                 "test7 - on  out border of this month",
		UserID:                "USER0",
		Description:           "",
		DateStart:             std.Add(743 * time.Hour),
		DateStop:              std.Add(750 * time.Hour),
		EventMessageTimeDelta: emtd,
	}
	events[8] = storage.Event{
		ID:                    0,
		Title:                 "test8 - in next month",
		UserID:                "USER0",
		Description:           "",
		DateStart:             std.Add(755 * time.Hour),
		DateStop:              std.Add(760 * time.Hour),
		EventMessageTimeDelta: emtd,
	}
	events[9] = storage.Event{
		ID:                    0,
		Title:                 "test9 - in previous day",
		UserID:                "USER0",
		Description:           "",
		DateStart:             std.Add(-20 * time.Hour),
		DateStop:              std.Add(-15 * time.Hour),
		EventMessageTimeDelta: emtd,
	}
	for _, curEvent := range events {
		_, err := s.CreateEvent(ctx, logger, curEvent)
		require.NoError(t, err)
	}
}

func TestStoragePositiveCreateEvent(t *testing.T) {
	logger, _ := logger.New("debug")
	s := initStorageInMemory(t)
	events := make([]storage.Event, 2)
	ctx := context.Background()
	std := time.Date(2023, 3, 27, 0, 0, 0, 1, time.Local)
	emtd := 4 * time.Hour
	events[0] = storage.Event{
		ID:                    0,
		Title:                 "test0 - baseDateTime",
		UserID:                "USER0",
		Description:           "",
		DateStart:             std.Add(5 * time.Hour),
		DateStop:              std.Add(7 * time.Hour),
		EventMessageTimeDelta: emtd,
	}
	events[1] = storage.Event{
		ID:                    0,
		Title:                 "test1 - in this day",
		UserID:                "USER0",
		Description:           "",
		DateStart:             std.Add(9 * time.Hour),
		DateStop:              std.Add(11 * time.Hour),
		EventMessageTimeDelta: emtd,
	}

	for _, curEvent := range events {
		t.Run("PositiveCreateEvent", func(t *testing.T) {
			curEvent := curEvent
			t.Parallel()
			id, err := s.CreateEvent(ctx, logger, curEvent)
			require.NoError(t, err)
			_, err = s.GetEvent(ctx, logger, id)
			require.NoError(t, err)
		})
	}
	err := s.Close(context.Background(), logger)
	require.NoError(t, err)
}

func TestStoragePositiveUpdateEvent(t *testing.T) {
	logger, _ := logger.New("debug")
	s := initStorageInMemory(t)
	ctx := context.Background()
	createTestEventPack(t, s)

	for i := 0; i < 2; i++ {
		t.Run("PositiveUpdateEvent", func(t *testing.T) {
			i := i
			t.Parallel()
			tEvent, err := s.GetEvent(ctx, logger, i)
			require.NoError(t, err)
			tEvent.Title = "testUpdated"
			err = s.UpdateEvent(ctx, logger, tEvent)
			require.NoError(t, err)
			testEvent, err := s.GetEvent(ctx, logger, i)
			require.NoError(t, err)
			require.Truef(t, testEvent.Title == "testUpdated", "event not update: ", testEvent.Title)
		})
	}

	err := s.Close(context.Background(), logger)
	require.NoError(t, err)
}

func TestStoragePositiveDeleteEvent(t *testing.T) {
	logger, _ := logger.New("debug")
	s := initStorageInMemory(t)
	ctx := context.Background()
	createTestEventPack(t, s)
	for i := 0; i < 2; i++ {
		t.Run("PositiveDeleteEvent", func(t *testing.T) {
			i := i
			err := s.DeleteEvent(ctx, logger, i)
			require.NoError(t, err)
			_, err = s.GetEvent(ctx, logger, 0)
			require.Truef(t, errors.Is(err, storage.ErrNoRecord), "actual error %q", err)
		})
	}
	err := s.Close(context.Background(), logger)
	require.NoError(t, err)
}

func TestStorage(t *testing.T) {
	logger, _ := logger.New("debug")
	t.Run("PositiveInit", func(t *testing.T) {
		s := New()
		err := s.Init(context.Background(), logger, nil)
		require.NoError(t, err)
		err = s.Close(context.Background(), logger)
		require.NoError(t, err)
	})

	t.Run("PositiveGetEvent", func(t *testing.T) {
		s := initStorageInMemory(t)
		ctx := context.Background()
		createTestEventPack(t, s)

		testEvent, err := s.GetEvent(ctx, logger, 0)
		require.NoError(t, err)
		require.Truef(t, testEvent.Title == "test0 - baseDateTime", "bad event", testEvent.Title)
		err = s.Close(context.Background(), logger)
		require.NoError(t, err)
	})

	t.Run("PositiveGetListEventsonDayByDay", func(t *testing.T) {
		s := initStorageInMemory(t)
		ctx := context.Background()
		createTestEventPack(t, s)
		testEventMap := make(map[int]bool)
		testEventMap[0] = false
		testEventMap[1] = false
		testEventMap[4] = false
		testEventMap[5] = false
		testtime := time.Date(2023, 3, 27, 12, 0, 0, 1, time.Local)
		testEvents, err := s.GetListEventsonDayByDay(ctx, logger, testtime)
		require.NoError(t, err)
		for _, curEvent := range testEvents {
			_, ok := testEventMap[curEvent.ID]
			require.Truef(t, ok, "bad event list(GetListEventsonDayByDay): ", testEvents)
		}
		err = s.Close(context.Background(), logger)
		require.NoError(t, err)
	})
	t.Run("PositiveGetListEventsOnWeekByDay", func(t *testing.T) {
		s := initStorageInMemory(t)
		ctx := context.Background()
		createTestEventPack(t, s)
		testEventMap := make(map[int]bool)
		testEventMap[0] = false
		testEventMap[1] = false
		testEventMap[2] = false
		testEventMap[4] = false
		testEventMap[5] = false
		testEventMap[6] = false
		testtime := time.Date(2023, 3, 27, 12, 0, 0, 1, time.Local)
		testEvents, err := s.GetListEventsonDayByDay(ctx, logger, testtime)
		require.NoError(t, err)
		for _, curEvent := range testEvents {
			_, ok := testEventMap[curEvent.ID]
			require.Truef(t, ok, "bad event list(GetListEventsOnWeekByDay): ", testEvents)
		}
		err = s.Close(context.Background(), logger)
		require.NoError(t, err)
	})
	t.Run("PositiveGetListEventsOnMonthByDay", func(t *testing.T) {
		s := initStorageInMemory(t)
		ctx := context.Background()
		createTestEventPack(t, s)
		testEventMap := make(map[int]bool)
		testEventMap[0] = false
		testEventMap[1] = false
		testEventMap[2] = false
		testEventMap[3] = false
		testEventMap[4] = false
		testEventMap[5] = false
		testEventMap[6] = false
		testEventMap[7] = false
		testtime := time.Date(2023, 3, 27, 12, 0, 0, 1, time.Local)
		testEvents, err := s.GetListEventsonDayByDay(ctx, logger, testtime)
		require.NoError(t, err)
		for _, curEvent := range testEvents {
			_, ok := testEventMap[curEvent.ID]
			require.Truef(t, ok, "bad event list(GetListEventsOnMonthByDay): ", testEvents)
		}
		err = s.Close(context.Background(), logger)
		require.NoError(t, err)
	})
	t.Run("NegativeCreateEventDateBusy", func(t *testing.T) {
		s := initStorageInMemory(t)
		ctx := context.Background()
		createTestEventPack(t, s)
		std := time.Date(2023, 3, 27, 0, 0, 0, 1, time.Local)
		emtd := 4 * time.Hour
		tEvent := storage.Event{
			ID:                    0,
			Title:                 "test777 - testDateBusy",
			UserID:                "USER0",
			Description:           "",
			DateStart:             std.Add(9 * time.Hour),
			DateStop:              std.Add(11 * time.Hour),
			EventMessageTimeDelta: emtd,
		}
		_, err := s.CreateEvent(ctx, logger, tEvent)

		require.Truef(t, errors.Is(err, storage.ErrDateBusy), "actual error %q", err)
		err = s.Close(context.Background(), logger)
		require.NoError(t, err)
	})
	t.Run("NegativeUpdateEventBadID", func(t *testing.T) {
		s := initStorageInMemory(t)
		ctx := context.Background()
		createTestEventPack(t, s)
		uEvent := storage.Event{}
		uEvent.Title = messageTestUpdated
		uEvent.ID = 25
		err := s.UpdateEvent(ctx, logger, uEvent)
		require.Truef(t, errors.Is(err, storage.ErrNoRecord), "actual error %q", err)
		err = s.Close(context.Background(), logger)
		require.NoError(t, err)
	})

	t.Run("NegativeDeleteEvent", func(t *testing.T) {
		s := initStorageInMemory(t)
		ctx := context.Background()
		createTestEventPack(t, s)
		err := s.DeleteEvent(ctx, logger, 25)
		require.Truef(t, errors.Is(err, storage.ErrNoRecord), "actual error %q", err)
		err = s.Close(context.Background(), logger)
		require.NoError(t, err)
	})
}
