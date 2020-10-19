# appointy-go-api

0.Run the main.go file
1.This is an API developed in GO with MongoDB. 
2.This API schedules a meeting given id,title,participants and other details, it has been implemented by using a POST method of net/http library of GO lang.
3.One can get the meeting details by providing the meeting ID.
4.It can also return the list of meetings arranged within a given time range.
5.Finally it can return all the meetings in which a particular participant is present.

6.Dependencies:

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

7.Constraints Satisfaction:
  7.1 Meetings for a particular participant should not be overlapped. Here it is achived using the functions checkValidity and checkParticipantAvailability.
  7.2 This API is thread-safe so there is not any chance of arriving at a RACE CONDITION. It is implemented here by using sync.lock
  7.3 Use API Pagination- This has been implemented by taking limit in GET request URL and implementing the logic in the function handler.
  7.4 Unit Testing-
 
8.Outputs:
All output screenshots are available in output folder.
