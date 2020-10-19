package main

//importing libraries
import(
	"strconv"
	"fmt"
	"sync"
	"log"
	"os"
	"net/http"
	"time"
	"strings"
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"

) 

//model for meeting
type Meeting struct{
	ID string `json:"id,omitempty" bson:"id,omitempty"`
	Title string `json:"title,omitempty" bson:"title,omitempty"`
	Participants []Participant `json:"participants" bson:"participants"`
	StartTime time.Time `json:"startTime" bson:"startTime"`
	EndTime time.Time `json:"endTime" bson:"endTime"`
	CreationTimestamp time.Time `json:"creationTimestamp" bson:"creationTimestamp"`

}

//model for meeting participants
type Participant struct{
	Name string `json:"name" bson:"name"`
	Email string `json:"email" bson:"email"`
	RSVP string `json:"rsvp" bson:"rsvp"`
}

//temporary model for taking input from GET request URL
type TempStruct struct {
	ID           string        `json:"id" bson:"id"`
	Title        string        `json:"title" bson:"title"`
	Participants []Participant `json:"participants" bson:"participants"`
	StartTime    string        `json:"startTime" bson:"startTime"`
	EndTime      string        `json:"endTime" bson:"endTime"`
}

//function to make database connection
func connectDatabase(){
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user :=os.Getenv("user")
	pass :=os.Getenv("pass")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
  	"mongodb+srv://"+user+":"+pass+"@cluster0.3uzwj.mongodb.net/appointy?retryWrites=true&w=majority",
	))
	if err != nil { log.Fatal(err) }
	appointyDatabase := client.Database("appointy")
	meetingCollection = appointyDatabase.Collection("meeting")
	fmt.Println("connected to database")
	return
}

//function to convert time in string format to time format
func strToTime(timeString string) (time.Time) {
	layout := "02-01-2006 03:04:05 PM"
	t, err := time.Parse(layout, timeString)
	if err != nil {
		log.Printf("Error while parsing time, Reason %v\n", err.Error())
		return time.Now()
	}
	return t
}

//checks whether the meeting is valid based on constraints
func checkValidity(participants []Participant, meetingStart time.Time, meetingEnd time.Time) (bool, int, error) {
	var flag bool
	var err error
	for i, p := range participants {
		flag, err = checkParticipantAvailability(p.Email, meetingStart, meetingEnd)
		if err != nil {
			return false, -1, err
		}
		if flag == true {
			return true, i, nil
		}
	}
	return false, -1, nil
}

//check if participant already has a meeting in a particular time frame
func checkParticipantAvailability(email string, meetingStart time.Time, meetingEnd time.Time) (bool, error) {
	cur, err := meetingCollection.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
		return false, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var meeting Meeting
		err = cur.Decode(&meeting)
		if err != nil {
			log.Fatal(err)
			return false, err
		}

		for _, p := range meeting.Participants {
			if p.Email == email {
				if meetingStart.Before(meeting.EndTime) && meetingStart.After(meeting.StartTime) ||
					meetingEnd.After(meeting.StartTime) && meetingEnd.Before(meeting.EndTime) {
					if p.RSVP == "yes" {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
}

//functin handler to get the meeting based on given ID
func GetMeetingUsingID(w http.ResponseWriter,r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet{
		w.WriteHeader(405)
		w.Write([]byte("Method Not Allowed"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	parts := strings.Split(r.URL.String(), "/")
	id := parts[2]
	var result Meeting
	meetingCollection.FindOne(ctx, bson.M{"id": id}).Decode(&result)
	defer cancel()
	meeting,_:=json.Marshal(result)
	w.Write(meeting)
}

//function for multiple endpoint handling based on request method 
func MultipleEndpointFunction(w http.ResponseWriter,r *http.Request){
	
	//making our function thread using sync.lock
	lock.Lock()
	defer lock.Unlock()
	switch r.Method{
	case "GET":
		w.Header().Set("Content-Type", "application/json")
		query :=r.URL.Query()
		startT :=query.Get("start")
		endT :=query.Get("end")
		participant :=query.Get("participant")
		limitParam :=query.Get("limit") // limit parameter for paging purpose
		limit,_ :=strconv.Atoi(limitParam) //converting string to int
		if len(startT) > 0 && len(endT) > 0{
			
		start:=strToTime(startT)
		end:=strToTime(endT)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cursor, err := meetingCollection.Find(ctx, bson.D{})
		defer cancel()
    	// Find() method raised an error
    	if err != nil {
        	fmt.Println("Finding all documents ERROR:", err)
        	defer cursor.Close(ctx)

    	// If the API call was a success
    	} else {
			var meetings []Meeting
        	// iterate over docs using Next()
       		 for cursor.Next(ctx) {

            // declare a result BSON object
            var result Meeting
            err := cursor.Decode(&result)

            // If there is a cursor.Decode error
            if err != nil {
                fmt.Println("cursor.Next() error:", err)
                os.Exit(1)
               
            // If there are no cursor.Decode errors
            } else {
				if start.Before(result.StartTime) || start.Equal(result.StartTime) &&
					end.After(result.EndTime) || end.Equal(result.EndTime) {
						if len(meetings)<limit{   //API paging using limit parameter
							meetings = append(meetings, result)
						}
				}
            }
		}
		js, err := json.Marshal(meetings)
			if err != nil {
				log.Printf("Error while marshalling JSON, Reason %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write(js)	
		}
	}else if len(participant) > 0 {

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			cur, err := meetingCollection.Find(ctx, bson.D{})
			if err != nil {
				log.Fatal(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer cancel()

			var meetings []Meeting
			for cur.Next(ctx) {
				var meeting Meeting
				err = cur.Decode(&meeting)
				if err != nil {
					log.Fatal(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				for _, p := range meeting.Participants {
					if p.Email == participant {
						if len(meetings)<limit{   //API paging using limit parameter
							meetings = append(meetings, meeting)
						}
						break
					}
				}
			}

			js, err := json.Marshal(meetings)
			if err != nil {
				log.Printf("Error while marshalling JSON, Reason %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write(js)

		} else {
			log.Printf("Wrong GET Query called")
			http.Error(w, "Wrong GET Query called", http.StatusNotImplemented)
		}
	
	case "POST":
		w.Header().Set("Content-Type", "application/json")
		// var temp Meeting
		var temp TempStruct
		var meeting Meeting
		json.NewDecoder(r.Body).Decode(&temp)
		// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		meeting.ID=temp.ID
		meeting.Title=temp.Title
		meeting.Participants=temp.Participants
		meeting.StartTime=strToTime(temp.StartTime)
		meeting.EndTime=strToTime(temp.EndTime)
		meeting.CreationTimestamp= time.Now()
		flag, index, err := checkValidity(meeting.Participants, meeting.StartTime, meeting.EndTime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if flag == true {
			log.Printf("Participant %s already has a meeting in this time period", meeting.Participants[index].Name)
			http.Error(w, "Participant "+meeting.Participants[index].Name+" already has a meeting in this time period", http.StatusNotAcceptable)
			return
		}
		_, err = meetingCollection.InsertOne(ctx, meeting)
		if err != nil {
			log.Printf("Error while inserting data, Reason %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		js, err := json.Marshal(meeting)
		if err != nil {
			log.Printf("Error while marshalling JSON, Reason %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(js)	
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Can't find method requested"}`))
	}
}

//global variables
var client *mongo.Client
var meetingCollection *mongo.Collection
var ctx context.Context
var lock sync.Mutex

//main function
func main(){
	fmt.Println("starting server")
	connectDatabase()
	http.HandleFunc("/meetings",MultipleEndpointFunction)
	http.HandleFunc("/meeting/",GetMeetingUsingID)
	http.ListenAndServe(":8000",nil)
}
