package internalhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/config"
	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/server"
	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/storage"
	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

func TestMemStorage(t *testing.T) {
	suite.Run(t, new(MarshallingSuite))
}

type MarshallingSuite struct {
	suite.Suite
	testData []storage.Event
}

func (s *MarshallingSuite) BuildStubRequestWithEvents(events ...storage.Event) *http.Request {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	var err error
	if len(events) == 1 {
		// this is a workaround to force TestReceiveJSONHelperWithSingleEvent pass
		err = encoder.Encode(&events[0])
	} else {
		err = encoder.Encode(&events)
	}

	s.Require().NoError(err)

	stubRequest := httptest.NewRequest("", "/stub_url", buf)
	stubRequest.Header.Set("Content-Type", "application/json")
	return stubRequest
}

func (s *MarshallingSuite) SetupTest() {
	var testEvent storage.Event
	for i := 0; i < 10; i++ {
		err := faker.FakeData(&testEvent)
		s.Require().NoError(err, "error during fake event generation")
		// monotonic clock leads to fail during require.Equal >_<
		testEvent.StartTime = testEvent.StartTime.Truncate(time.Nanosecond).Local()
		testEvent.EndTime = testEvent.EndTime.Truncate(time.Nanosecond).Local()
		s.testData = append(s.testData, testEvent)
	}
}

func (s *MarshallingSuite) TestEmptyMarshalling() {
	var testEvent storage.Event
	var resEvent storage.Event

	r := httptest.NewRecorder()
	err := sendJSON(r, testEvent)
	s.Require().NoError(err)

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&resEvent)
	s.Require().NoError(err)
	s.Require().True(testEvent.IsEqual(resEvent))
}

func (s *MarshallingSuite) TestEmptyUnmarshalling() {
	var testEvent storage.Event
	var resEvent storage.Event

	stubRequest := s.BuildStubRequestWithEvents(testEvent)
	err := receiveJSON(stubRequest, &resEvent)
	s.Require().NoError(err)
	s.Require().True(testEvent.IsEqual(resEvent))
}

func (s *MarshallingSuite) TestSendJSONHelperWithSingleEvent() {
	var resData storage.Event

	r := httptest.NewRecorder()
	err := sendJSON(r, &s.testData[0])
	s.Require().NoError(err)

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&resData)
	s.Require().NoError(err)
	s.Require().True(s.testData[0].IsEqual(resData))
}

func (s *MarshallingSuite) TestSendJSONHelperWithEventSlice() {
	var resData []storage.Event

	r := httptest.NewRecorder()
	err := sendJSON(r, s.testData)
	s.Require().NoError(err)

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&resData)
	s.Require().NoError(err)
	s.Require().Len(resData, len(s.testData))
	s.Require().True(IsEqual(s.testData, resData))
}

func (s *MarshallingSuite) TestReceiveJSONHelperWithSingleEvent() {
	var resData storage.Event
	stubRequest := s.BuildStubRequestWithEvents(s.testData[0])
	err := receiveJSON(stubRequest, &resData)
	s.Require().NoError(err)
	s.Require().True(s.testData[0].IsEqual(resData))
}

func (s *MarshallingSuite) TestReceiveJSONHelperWithEventSlice() {
	var resData []storage.Event
	stubRequest := s.BuildStubRequestWithEvents(s.testData...)
	err := receiveJSON(stubRequest, &resData)
	s.Require().NoError(err)
	s.Require().Len(resData, len(s.testData))
	s.Require().True(IsEqual(s.testData, resData))
}

func TestHTTPApi(t *testing.T) {
	suite.Run(t, new(HTTPApiSuite))
}

type HTTPApiSuite struct {
	suite.Suite
	testSlice      []storage.Event
	testEvent      storage.Event
	testCreateData CreateEventData
	ctl            *gomock.Controller
	mockedApp      *server.MockApplication
	testServer     *httptest.Server
	ctx            context.Context
	cancelFunc     context.CancelFunc
}

func (s *HTTPApiSuite) SetupSuite() {
	s.ctx, s.cancelFunc = context.WithTimeout(context.Background(), time.Second*5)
	s.ctl = gomock.NewController(s.T())
	s.mockedApp = server.NewMockApplication(s.ctl)

	// faking test only CreateEventData
	err := faker.FakeData(&s.testCreateData)
	// monotonic clock leads to fail during require.Equal >_<
	s.testCreateData.StartTime = s.testCreateData.StartTime.Truncate(time.Nanosecond).Local()
	s.testCreateData.EndTime = s.testCreateData.EndTime.Truncate(time.Nanosecond).Local()
	s.Require().NoError(err, "error during fake create data generation")

	// faking test only storage.Event
	err = faker.FakeData(&s.testEvent)
	// monotonic clock leads to fail during require.Equal >_<
	s.testEvent.StartTime = s.testEvent.StartTime.Truncate(time.Nanosecond).Local()
	s.testEvent.EndTime = s.testEvent.EndTime.Truncate(time.Nanosecond).Local()
	s.Require().NoError(err, "error during fake event generation")

	// faking test only []storage.Event
	var testEvent storage.Event
	for i := 0; i < 10; i++ {
		err := faker.FakeData(&testEvent)
		s.Require().NoError(err, "error during fake event generation")
		// monotonic clock leads to fail during require.Equal >_<
		testEvent.StartTime = testEvent.StartTime.Truncate(time.Nanosecond).Local()
		testEvent.EndTime = testEvent.EndTime.Truncate(time.Nanosecond).Local()
		s.testSlice = append(s.testSlice, testEvent)
	}

	// for router tests purposes creating httptest.Server
	api := NewHTTPApi(config.HTTPApiConfig{
		Port: 8888,
	}, s.mockedApp)
	s.testServer = httptest.NewServer(api.server.Handler)
}

func (s *HTTPApiSuite) TearDownSuite() {
	defer s.testServer.Close()
	defer s.cancelFunc()
}

func (s *HTTPApiSuite) TestAddEventHandler() {
	marshal, err := json.Marshal(s.testCreateData)
	s.Require().NoError(err)
	r := httptest.NewRequest("POST", "localhost:8080/calendar/add", bytes.NewBuffer(marshal))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.mockedApp.EXPECT().CreateEvent(
		gomock.Any(),
		s.testCreateData.Title,
		gomock.Any(),
		gomock.Any(),
		s.testCreateData.Description,
		s.testCreateData.OwnerID,
	).Return(storage.Event{
		ID:          faker.UUIDHyphenated(),
		Title:       s.testCreateData.Title,
		StartTime:   s.testCreateData.StartTime,
		EndTime:     s.testCreateData.EndTime,
		Description: s.testCreateData.Description,
		OwnerID:     s.testCreateData.OwnerID,
	}, nil)
	service := &Service{app: s.mockedApp}
	service.AddEventHandler(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)
	s.Require().Equal("application/json", resp.Header.Get("Content-Type"))

	var resEvent storage.Event
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&resEvent)

	s.Require().NoError(err)
	s.Require().Equal(s.testCreateData.Title, resEvent.Title)
	s.Require().True(s.testCreateData.StartTime.Equal(resEvent.StartTime))
	s.Require().True(s.testCreateData.EndTime.Equal(resEvent.EndTime))
	s.Require().Equal(s.testCreateData.Description, resEvent.Description)
	s.Require().Equal(s.testCreateData.OwnerID, resEvent.OwnerID)
}

func (s *HTTPApiSuite) TestUpdateEventHandler() {
	marshal, err := json.Marshal(s.testEvent)
	s.Require().NoError(err)
	r := httptest.NewRequest("POST", "localhost:8080/calendar/update", bytes.NewBuffer(marshal))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.mockedApp.EXPECT().UpdateEvent(
		gomock.Any(),
		eventsMatcher{s.testEvent},
	).Return(nil)
	service := &Service{app: s.mockedApp}
	service.UpdateEventHandler(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)
	s.Require().Equal("application/json", resp.Header.Get("Content-Type"))

	var resEvent storage.Event
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&resEvent)

	s.Require().NoError(err)
	s.Require().True(s.testEvent.IsEqual(resEvent))
}

func (s *HTTPApiSuite) OtherHandlers() {
	s.T().Skip()
	// well, i didn't find a way to test service.DeleteEventHandler() service.FindEventsHandler() handlers standalone
	// it's because inside it's methods a mux.Vars() invocation occurred.
	// mux.Vars() invocation requires request.Context to contain gorilla/mux parsed path variables
	// and i didn't find a good approach how to do it :(
}

func (s *HTTPApiSuite) TestAddEvent() {
	marshal, err := json.Marshal(s.testCreateData)
	s.Require().NoError(err)
	r, err := http.NewRequestWithContext(s.ctx, "POST", s.testServer.URL+"/calendar/add", bytes.NewBuffer(marshal))
	s.Require().NoError(err)
	r.Header.Set("Content-Type", "application/json")

	s.mockedApp.EXPECT().CreateEvent(
		gomock.Any(),
		s.testCreateData.Title,
		gomock.Any(),
		gomock.Any(),
		s.testCreateData.Description,
		s.testCreateData.OwnerID,
	).Return(storage.Event{
		ID:          faker.UUIDHyphenated(),
		Title:       s.testCreateData.Title,
		StartTime:   s.testCreateData.StartTime,
		EndTime:     s.testCreateData.EndTime,
		Description: s.testCreateData.Description,
		OwnerID:     s.testCreateData.OwnerID,
	}, nil)

	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Do(r)
	s.Require().NoError(err)
	defer func() {
		err := resp.Body.Close()
		s.Require().NoError(err)
	}()

	s.Require().Equal(http.StatusOK, resp.StatusCode)
	s.Require().Equal("application/json", resp.Header.Get("Content-Type"))

	var resEvent storage.Event
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&resEvent)

	s.Require().NoError(err)
	s.Require().Equal(s.testCreateData.Title, resEvent.Title)
	s.Require().True(s.testCreateData.StartTime.Equal(resEvent.StartTime))
	s.Require().True(s.testCreateData.EndTime.Equal(resEvent.EndTime))
	s.Require().Equal(s.testCreateData.Description, resEvent.Description)
	s.Require().Equal(s.testCreateData.OwnerID, resEvent.OwnerID)
}

func (s *HTTPApiSuite) TestUpdateEvent() {
	marshal, err := json.Marshal(s.testEvent)
	s.Require().NoError(err)
	r, err := http.NewRequestWithContext(s.ctx, "POST", s.testServer.URL+"/calendar/update", bytes.NewBuffer(marshal))
	s.Require().NoError(err)
	r.Header.Set("Content-Type", "application/json")

	s.mockedApp.EXPECT().UpdateEvent(
		gomock.Any(),
		eventsMatcher{s.testEvent},
	).Return(nil)

	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Do(r)
	s.Require().NoError(err)
	defer func() {
		s.Require().NoError(resp.Body.Close())
	}()

	s.Require().Equal(http.StatusOK, resp.StatusCode)
	s.Require().Equal("application/json", resp.Header.Get("Content-Type"))

	var resEvent storage.Event
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&resEvent)

	s.Require().NoError(err)
	s.Require().True(s.testEvent.IsEqual(resEvent))
}

func (s *HTTPApiSuite) TestDeleteEvent() {
	request, err := http.NewRequestWithContext(s.ctx, "POST", s.testServer.URL+"/calendar/delete/TEST_EVENT_ID", nil)
	s.Require().NoError(err)

	s.mockedApp.EXPECT().DeleteEvent(
		gomock.Any(),
		gomock.Eq("TEST_EVENT_ID"),
	).Return(nil)

	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Do(request)
	s.Require().NoError(err)
	defer func() {
		s.Require().NoError(resp.Body.Close())
	}()
	s.Require().Equal(http.StatusOK, resp.StatusCode)
}

func (s *HTTPApiSuite) TestFindEvents() {
	request, err := http.NewRequestWithContext(s.ctx, "GET", s.testServer.URL+"/calendar/find/day/2021/08/25", nil)
	s.Require().NoError(err)

	s.mockedApp.EXPECT().ListDayEvents(
		gomock.Any(),
		gomock.Any(),
	).Return(s.testSlice, nil)

	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Do(request)
	s.Require().NoError(err)
	defer func() {
		s.Require().NoError(resp.Body.Close())
	}()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var result []storage.Event
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	s.Require().NoError(err)
	s.Require().True(IsEqual(s.testSlice, result))
}

func IsEqual(s1 []storage.Event, s2 []storage.Event) bool {
	if s1 == nil && s2 == nil {
		return true
	}
	if s1 == nil || s2 == nil {
		return false
	}
	if len(s1) != len(s2) {
		return false
	}

	for i, e := range s1 {
		if !e.IsEqual(s2[i]) {
			return false
		}
	}
	return true
}

type eventsMatcher struct {
	storage.Event
}

func (m eventsMatcher) Matches(x interface{}) bool {
	e, ok := x.(storage.Event)
	if !ok {
		return false
	}
	return m.Event.IsEqual(e)
}

func (m eventsMatcher) String() string {
	return fmt.Sprintf("is equal to %+v", m.Event)
}
