package main

import(
	"net/http"
    "net/http/httptest"
    "testing"
)


func TestGetMeetingUsingID(t *testing.T) {
    req, err := http.NewRequest("GET", "/meeting/1", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(GetMeetingUsingID)
    handler.ServeHTTP(rr, req)
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v",
            status, http.StatusOK)
    }
    // Check the response body is what we expect.
    expected := `{"_id":{"$oid":"5f8d13ce1bb80eb43fe72b51"},"id":"1","title":"meeting 1","participants":[{"name":"raghav","email":"raghav@gmail.com","rsvp":"no"},{"name":"eva","email":"eva@gmail.com","rsvp":"yes"}],"startTime":{"$date":{"$numberLong":"1603101600000"}},"endTime":{"$date":{"$numberLong":"1603108800000"}},"creationTimestamp":{"$date":{"$numberLong":"1603081166427"}}}`
    if rr.Body.String() != expected {
        t.Errorf("handler returned unexpected body: got %v want %v",
            rr.Body.String(), expected)
    }
}
func TestMultipleEndpointFunction(t *testing.T){
	req, err := http.NewRequest("POST", "/meetings", nil)
    if err != nil {
        t.Fatal(err)
	}
	rr := httptest.NewRecorder()
    handler := http.HandlerFunc(GetMeetingUsingID)
    handler.ServeHTTP(rr, req)
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v",
            status, http.StatusOK)
    }
}
