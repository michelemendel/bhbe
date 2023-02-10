package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/gorilla/mux"
	"github.com/michelemendel/bhbe/algolia"
	"github.com/michelemendel/bhbe/redis"
	"github.com/michelemendel/bhbe/utils"
	"go.uber.org/zap"
)

// https://grciuta.medium.com/server-sent-events-with-go-server-and-client-sides-6812dca45c7

var lg *zap.SugaredLogger

func init() {
	lg = utils.Log()
}

const (
	allowedHeaders = "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization,X-CSRF-Token"
)

// UUID is the unique identifier of the client
type connection struct {
	clientUuid      string //todo rename to UUID
	algoliaObjectId string //todo remove
	remoteAddr      string
	writer          http.ResponseWriter
	flusher         http.Flusher
	requestCtx      context.Context //Not sure why this is needed
}

type serverCtx struct {
	connections map[string]*connection // The key is the clientUuid of the client
	redCtx      *redis.RedisCtx
	//todo remove aXlgoliaIndex
	algoliaIndex *search.Index
	mutex        *sync.Mutex
}

// todo remove searchIndex
func newServerCtx(searchIndex *search.Index) *serverCtx {
	return &serverCtx{
		connections: map[string]*connection{},
		redCtx:      redis.InitRedisClient(),
		//todo remove aXlgoliaIndex
		algoliaIndex: searchIndex,
		mutex:        &sync.Mutex{},
	}
}

type messageType string

const (
	broadcast         messageType = "broadcast"
	broadcastToSender messageType = "broadcasttosender"
	reply             messageType = "reply"
	replyToSender     messageType = "replytosender"
	msgError          messageType = "error"
)

// Broadcast and reply message
type message struct {
	Error            string       `json:"error"`
	MessageType      messageType  `json:"messagetype"`
	BroadcastUUID    string       `json:"broadcastuuid"`
	RepliesUUID      string       `json:"repliesuuid"`
	FromUUID         string       `json:"fromuuid"`
	ToUUID           string       `json:"touuid"`
	FromUsername     string       `json:"fromusername"`
	ToUsername       string       `json:"tousername"`
	FromLoc          utils.Coords `json:"fromloc"`
	ToLoc            utils.Coords `json:"toloc"`
	TargetLoc        utils.Coords `json:"targetloc"`
	Radius           float64      `json:"radius"`
	DistanceToTarget float64      `json:"distancetotarget"`
	DistanceToSender float64      `json:"distancetosender"`
	MsgRows          []string     `json:"msgrows"`
	NofRecipients    int          `json:"nofrecipients"`
	Timer            int          `json:"timer"`
	CreatedAt        string       `json:"createdat"`
}

// --------------------------------------------------------------------------------

// todo remove aXlgoliaIndex
func StartApiServer(hostAddr, port string, algoliaIndex *search.Index) {
	//todo remove aXlgoliaIndex
	sCtx := newServerCtx(algoliaIndex)
	router := mux.NewRouter()

	// ----------------------------------------
	// Operational endpoints
	// curl -X GET "http://localhost:8588/cleardb"
	router.HandleFunc("/cleardb", sCtx.clearServerCtxHandler).Methods("GET")
	// curl -X GET "https://192.168.39.113:8588/info" -k
	router.HandleFunc("/info", sCtx.GetServerCtxInfoHandler).Methods("GET")
	// todo add endpoint to get info from redis. Maybe combine with the above one.

	// ----------------------------------------
	// Client endpoints
	// SSE
	router.HandleFunc("/events/{clientUuid}", sCtx.SSEHandler).Methods("GET")
	// curl -X GET "http://localhost:8588/register/123?lat=1.23&lng=4.56"
	router.HandleFunc("/register/", sCtx.RegisterHandler).Methods("GET")
	router.HandleFunc("/register/{clientUuid}", sCtx.RegisterHandler).Methods("GET")
	// curl -X GET "http://localhost:8588/updategeo/123?lat=1.23&lng=4.56"
	router.HandleFunc("/updategeo/{clientUuid}", sCtx.UpdateGeoLocationHandler).Methods("GET")
	// curl -X POST 'https://192.168.39.113:8588/send/123' -k --data-raw 'location=Sao+Paulo%2C+-23.5%2C-46.6&message=xyz&timer=1'
	router.HandleFunc("/broadcast/{clientUuid}", sCtx.BroadcastHandler).Methods("POST")
	router.HandleFunc("/reply/{clientUuid}", sCtx.ReplyHandler).Methods("POST")

	lg.Infof("API Server started on %s:%s...", hostAddr, port)
	certPath := os.Getenv("CERT_PATH")
	http.ListenAndServeTLS(":"+port, certPath+"/cert.pem", certPath+"/key.pem", router)
}

// Handler for getting a UUID and registering a client
func (sCtx *serverCtx) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	clientUuid := mux.Vars(r)["clientUuid"]

	if clientUuid == "" {
		clientUuid = utils.GenerateUUID()
		lg.Infof("[Register] Creating a new clientUuid: %s", clientUuid)
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
	fmt.Fprintf(w, "%s\n", clientUuid)
}

// --------------------------------------------------------------------------------
func (sCtx *serverCtx) UpdateGeoLocationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)

	clientUUID := mux.Vars(r)["clientUuid"]
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	mylat := coordAsFloat(latStr)
	myLng := coordAsFloat(lngStr)
	fromUsername := r.URL.Query().Get("fromusername")
	if fromUsername == "" {
		fromUsername = "anonymous"
	}

	conn := sCtx.getConnection(clientUUID)
	if conn == nil {
		lg.Infof("[UpdateGeoLocation] No conn for clientUuid:%s", clientUUID)
		fmt.Fprintf(w, "%s\n", "")
		return
	}

	//TODO: mendel change to redis
	lg.Infof("[UpdateGeoLocation] clientUuid:%s, objectId:%v, lat:%v, lng:%v, fromUsername:%s", clientUUID, conn.algoliaObjectId, mylat, myLng, fromUsername)
	//TODO: mendel change to redis
	algolia.Update(sCtx.algoliaIndex, conn.algoliaObjectId, string(clientUUID), fromUsername, mylat, myLng)

	// todo move to top
	clientPrefix := "client:"
	sCtx.redCtx.CreateClient(clientPrefix, fromUsername)
	sCtx.redCtx.UpsertGeo(redis.UUID(clientUUID), mylat, myLng)

	fmt.Fprintf(w, "%s\n", clientUUID)
}

// --------------------------------------------------------------------------------
func sseHeaders(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Retry", "5000")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
}

// --------------------------------------------------------------------------------
func (sCtx *serverCtx) SSEHandler(rw http.ResponseWriter, req *http.Request) {
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	requestContext := req.Context()
	clientUuid := mux.Vars(req)["clientUuid"]
	conn := sCtx.resolveConnection(clientUuid)
	conn.remoteAddr = req.RemoteAddr
	conn.writer = rw
	conn.flusher = flusher
	conn.requestCtx = requestContext

	sseHeaders(rw)
	lg.Infof("[SSEHandler] User %s has connected from %s (%v)", clientUuid, req.RemoteAddr, len(sCtx.connections))

	sCtx.sendOpenEvent(conn)

	defer func() {
		//todo change to redis
		lg.Infof("[SSEHandler] Removing clientID: %s, objectID: %s", conn.clientUuid, conn.algoliaObjectId)
		sCtx.removeConnection(conn)
	}()

	// Why is this needed?
	<-requestContext.Done()
}

func (sCtx *serverCtx) sendOpenEvent(conn *connection) {
	if conn == nil {
		lg.Errorf("[sendOpenEvent] Connection is nil")
		return
	}

	messageBytes := []byte(fmt.Sprintf("event: open\ndata: %s\n\n", "Connection opened"))
	_, err := conn.writer.Write(messageBytes)
	if err != nil {
		lg.Errorf("[sendOpenEvent] Error getting writer %s: %v", conn.clientUuid, err)
		sCtx.removeConnection(conn)
	}
	conn.flusher.Flush()
}

// Handler to get db info
func (sCtx *serverCtx) GetServerCtxInfoHandler(w http.ResponseWriter, r *http.Request) {
	msg := sCtx.getServerConnections()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, "Connections: (%d)\n%v\n", len(sCtx.connections), strings.Join(msg, "\n"))
}

func (sCtx *serverCtx) getServerConnections() []string {
	msg := []string{}
	for _, conn := range sCtx.connections {
		msg = append(msg, fmt.Sprintf("UUID:%s, remoteAddr:%s", conn.clientUuid, conn.remoteAddr))
	}
	return msg
}

func (sCtx *serverCtx) clearServerCtxHandler(w http.ResponseWriter, r *http.Request) {
	sCtx.clearServerCtxAndAlgoliaIndex()
}

// TODO: mendel change to redis
func (sCtx *serverCtx) clearServerCtxAndAlgoliaIndex() {
	sCtx.mutex.Lock()
	defer sCtx.mutex.Unlock()
	sCtx.connections = make(map[string]*connection)
	//TODO: mendel change to use redis
	algolia.ClearIndex(sCtx.algoliaIndex)
}

// --------------------------------------------------------------------------------
func (sCtx *serverCtx) BroadcastHandler(w http.ResponseWriter, r *http.Request) {
	fromUuid := mux.Vars(r)["clientUuid"]

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "[BroadcastHandler] Error parsing form data", http.StatusBadRequest)
		return
	}

	// --------------------------------------------------------------------------------
	// User's nickname
	fromUsername := r.FormValue("fromusername")
	if fromUsername == "" {
		lg.Infof("[BroadcastHandler] No nickname specified, using John Doe")
		fromUsername = "anonymous"
	}

	// --------------------------------------------------------------------------------
	// The device's geo location
	fromLocLatStr := r.FormValue("myloclat")
	fromLocLngStr := r.FormValue("myloclng")
	if fromLocLatStr == "" || fromLocLngStr == "" {
		lg.Infof("[BroadcastHandler] No mylocationlat or mylocationlng specified, using (0,0)")
		fromLocLatStr = "0"
		fromLocLngStr = "0"
	}
	fromLocLat := coordAsFloat(fromLocLatStr)
	fromLocLng := coordAsFloat(fromLocLngStr)

	// ----------------------------------------
	// User's selected geo location
	targetLocLatStr := r.FormValue("targetloclat")
	targetLocLngStr := r.FormValue("targetloclng")
	if targetLocLatStr == "" || targetLocLngStr == "" {
		lg.Infof("[BroadcastHandler] No targetLocLat or targetLocLng specified, using MyLoc: (%s,%s)", fromLocLatStr, fromLocLngStr)
		targetLocLatStr = fromLocLatStr
		targetLocLngStr = fromLocLngStr
	}
	targetLocLat := coordAsFloat(targetLocLatStr)
	targetLocLng := coordAsFloat(targetLocLngStr)

	// ----------------------------------------
	broadcastUUID := utils.GenerateUUID()
	broadcastMessage := r.FormValue("broadcastmessage")
	if broadcastMessage == "" {
		broadcastMessage = ""
	}

	// ----------------------------------------
	timerStr := r.FormValue("timer")
	if timerStr == "" {
		// lg.Infof("[BroadcastHandler] No timer specified, using default (5 minutes)")
		timerStr = "5" //minutes
	}

	// minutes
	timer, err := strconv.Atoi(timerStr)
	if err != nil {
		lg.Errorf("[BroadcastHandler] Error converting timer to int, using default (5 minutes): %v", err)
		timer = 5 //minutes
	}

	// ----------------------------------------
	radiusStr := r.FormValue("radius")
	if radiusStr == "" {
		lg.Infof("[BroadcastHandler] No radius specified, using default (1 km)")
		radiusStr = "1" //km
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		lg.Errorf("[BroadcastHandler] Error converting radius to int using default (1 km): %v", err)
		radius = 1 //km
	}

	msgRows := strings.Split(broadcastMessage, "\n")
	fromLoc := utils.MakeCoords(fromLocLat, fromLocLng)
	targetLoc := utils.MakeCoords(targetLocLat, targetLocLng)

	// Broadcast message
	msg := message{
		Error:            "",
		MessageType:      broadcast,
		BroadcastUUID:    broadcastUUID,
		RepliesUUID:      "",
		FromUUID:         fromUuid,
		ToUUID:           "",
		FromUsername:     fromUsername,
		ToUsername:       "",
		FromLoc:          fromLoc,
		ToLoc:            utils.MakeCoords(0, 0),
		TargetLoc:        targetLoc,
		DistanceToTarget: -1,
		DistanceToSender: -1,
		MsgRows:          msgRows,
		NofRecipients:    -1, //This information is only for the broadcaster
		Radius:           radius,
		Timer:            timer,
	}

	// Send to all clients
	nofRecipients := sCtx.broadcast(msg)
	msg.NofRecipients = nofRecipients
	// Send back to the broadcaster
	msg.MessageType = broadcastToSender
	sCtx.sendToOne(msg)
}

// --------------------------------------------------------------------------------
func (sCtx *serverCtx) ReplyHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "[ReplyHandler] Error parsing form data", http.StatusBadRequest)
		return
	}

	// ----------------------------------------
	broadcastUUID := r.FormValue("replybroadcastuuid")
	if broadcastUUID == "" {
		lg.Errorf("[ReplyHandler] No broadcastUuid specified")
		return
	}

	repliesUUID := r.FormValue("repliesuuid")
	if repliesUUID == "" {
		lg.Infof("[ReplyHandler] No repliesUuid specified. Creating a new one")
		repliesUUID = utils.GenerateUUID()
	}

	// ----------------------------------------
	fromUUID := mux.Vars(r)["clientUuid"]
	if fromUUID == "" {
		lg.Errorf("[ReplyHandler] No fromUuid specified")
		return
	}
	toUUID := r.FormValue("replytouuid")
	if toUUID == "" {
		lg.Errorf("[ReplyHandler] No toUuid specified")
		return
	}

	// --------------------------------------------------------------------------------
	fromUsername := r.FormValue("replyfromusername")
	toUsername := r.FormValue("replytousername")
	if fromUsername == "" {
		lg.Infof("[ReplyHandler] No fromUserName (nickname) specified")
		fromUsername = "anonymous sender"
	}
	if toUsername == "" {
		lg.Infof("[ReplyHandler] No toUsername specified")
		toUsername = "anonymous receiver"
	}

	// --------------------------------------------------------------------------------
	// The device's geo location
	fromLocLatStr := r.FormValue("replymyloclat")
	fromLocLngStr := r.FormValue("replymyloclng")
	if fromLocLatStr == "" || fromLocLngStr == "" {
		lg.Infof("[ReplyHandler] No mylocationlat or mylocationlng specified, using (0,0)")
		fromLocLatStr = "0"
		fromLocLngStr = "0"
	}
	fromLocLat := coordAsFloat(fromLocLatStr)
	fromLocLng := coordAsFloat(fromLocLngStr)
	fromLoc := utils.MakeCoords(fromLocLat, fromLocLng)

	replyMessage := r.FormValue("replymessage")
	if replyMessage == "" {
		replyMessage = ""
	}

	msgRows := strings.Split(replyMessage, "\n")

	// Reply message
	replyMsg := message{
		Error:            "",
		MessageType:      reply,
		BroadcastUUID:    broadcastUUID,
		RepliesUUID:      repliesUUID,
		FromUUID:         fromUUID,
		ToUUID:           toUUID,
		FromUsername:     fromUsername,
		ToUsername:       toUsername,
		FromLoc:          fromLoc,
		ToLoc:            utils.MakeCoords(0, 0),
		TargetLoc:        utils.MakeCoords(0, 0), //Not relevant here
		Radius:           -1,                     //Not relevant here
		DistanceToTarget: -1,                     //Not relevant here
		DistanceToSender: -1,                     //Will be calculated below
		MsgRows:          msgRows,
		NofRecipients:    -1, //Not relevant here
		Timer:            -1, //Not relevant here
	}

	// Reply message
	replyToSenderMsg := message{
		Error:            "",
		MessageType:      replyToSender,
		BroadcastUUID:    broadcastUUID,
		RepliesUUID:      repliesUUID,
		FromUUID:         toUUID,
		ToUUID:           fromUUID,
		FromUsername:     toUsername,
		ToUsername:       fromUsername,
		FromLoc:          fromLoc,
		ToLoc:            fromLoc,
		TargetLoc:        utils.MakeCoords(0, 0), //Not relevant here
		Radius:           -1,                     //Not relevant here
		DistanceToTarget: -1,                     //Not relevant here
		DistanceToSender: -1,                     //Not relevant here
		MsgRows:          msgRows,
		NofRecipients:    -1, //Not relevant here
		Timer:            -1, //Not relevant here
	}

	// from->to
	err = sCtx.sendToOne(replyMsg)
	if err == nil {
		// from->from (i.e. back to sender)
		sCtx.sendToOne(replyToSenderMsg)
	}
}

// --------------------------------------------------------------------------------
func (sCtx *serverCtx) broadcast(msg message) int {
	// Note: Some of the recipients may not represent an active connection. This will be dealt with in the loop below

	//TODO: mendel change to redis SearchLocation
	// todo change msg.Radius to float64
	recipientsX := sCtx.redCtx.SearchLocation(redis.InitRedisClient().SearchLocationQuery(msg.TargetLoc.Lat, msg.TargetLoc.Lng, msg.Radius))
	fmt.Println("recipientsX: ", recipientsX)

	recipients := algolia.Recipients(sCtx.algoliaIndex, msg.TargetLoc.Lat, msg.TargetLoc.Lng, int(msg.Radius))

	fromConn := sCtx.connections[msg.FromUUID]
	var toConn *connection
	var toUUID string
	var toLat float64
	var toLng float64
	nofRecipients := 0

	for i, recipient := range recipients {
		if recipient == nil {
			continue
		}

		toUUID = (*recipient).ClientUuid
		toLat = (*recipient).Lat
		toLng = (*recipient).Lng

		toConn = sCtx.getConnection(toUUID)
		if toConn == nil {
			lg.Errorf("[broadcast] No connection for %s", toUUID)
			sCtx.removeConnection(toConn)
			continue
		}

		// Don't send to broadcaster. This is done separately. We want to count nof recipients and it's easier to do it this way.
		if toUUID == msg.FromUUID {
			continue
		}

		msg.ToUUID = toUUID
		msg.DistanceToTarget = utils.Distance(msg.TargetLoc.Lat, msg.TargetLoc.Lng, toLat, toLng)
		msg.DistanceToSender = utils.Distance(msg.FromLoc.Lat, msg.FromLoc.Lng, toLat, toLng)
		msg.CreatedAt = utils.StampTimeNow()

		lg.Infow(
			fmt.Sprintf("[Broadcast %d", i),
			"fromAddr:", fromConn.remoteAddr,
			"toAddr:", toConn.remoteAddr,
			"messageType:", msg.MessageType,
			"broadcastUUID:", msg.BroadcastUUID,
			"repliesUUID:", msg.RepliesUUID,
			"fromUUID:", msg.FromUUID,
			"toUUID:", toUUID,
			"fromUsername:", msg.FromUsername,
			"toUsername:", msg.ToUsername,
			"fromLoc:", fmt.Sprintf("%v,%v", msg.FromLoc.Lat, msg.FromLoc.Lng),
			"toLoc:", fmt.Sprintf("%v,%v", toLat, toLng),
			"targetLoc:", fmt.Sprintf("%v,%v", msg.TargetLoc.Lat, msg.TargetLoc.Lng),
			"DistanceToTarget:", msg.DistanceToTarget,
			"DistanceToSender:", msg.DistanceToSender,
			"message:", msg.MsgRows,
			"radius:", msg.Radius,
			"timer:", msg.Timer,
			"createdAt:", msg.CreatedAt,
		)

		message, err := json.Marshal(msg)
		if err != nil {
			message = []byte("{\"error\": \"There was a problem handling the broadcast message.\"}")
		}
		messageBytes := []byte(fmt.Sprintf("event: broadcast\ndata: %s\n\n", string(message)))

		_, err = toConn.writer.Write(messageBytes)
		if err != nil {
			lg.Errorf("[broadcast] Error getting writer %s: %v", toUUID, err)
			sCtx.removeConnection(toConn)
			continue
		}
		nofRecipients++
		toConn.flusher.Flush()
	}

	return nofRecipients
}

// --------------------------------------------------------------------------------
func (sCtx *serverCtx) sendToOne(msg message) error {
	eventType := "reply"
	msg.DistanceToSender = -1

	if msg.MessageType == broadcastToSender {
		eventType = "broadcast"
		msg.ToUUID = msg.FromUUID //Send back to sender
	} else if msg.MessageType == reply {

		// todo change to GetGeo
		recipientX := sCtx.redCtx.GetGeo(redis.UUID(msg.ToUUID))
		fmt.Println("recipientX: ", recipientX)
		recipient := algolia.Recipient(sCtx.algoliaIndex, msg.ToUUID)
		msg.DistanceToSender = utils.Distance(msg.FromLoc.Lat, msg.FromLoc.Lng, recipient.Lat, recipient.Lng)
	}

	fromConn := sCtx.getConnection(msg.FromUUID)
	toConn := sCtx.getConnection(msg.ToUUID)
	if fromConn == nil {
		lg.Infof("[sendToOne] No fromConn for %s", msg.ToUUID)
		sCtx.sendError(toConn, msg, "removebroadcast")
		return fmt.Errorf("[sendToOne] No fromConn for %s", msg.ToUUID)
	}
	if toConn == nil {
		lg.Infof("[sendToOne] No toConn for %s", msg.ToUUID)
		sCtx.sendError(fromConn, msg, "removebroadcast")
		return fmt.Errorf("[sendToOne] No toConn for %s", msg.ToUUID)
	}

	msg.CreatedAt = utils.StampTimeNow()

	lg.Infow("[Reply]",
		"eventType:", eventType,
		"fromAddr:", fromConn.remoteAddr,
		"toAddr:", toConn.remoteAddr,
		"messageType:", msg.MessageType,
		"broadcastUUID:", msg.BroadcastUUID,
		"repliesUUID:", msg.RepliesUUID,
		"fromUUID:", msg.FromUUID,
		"toUUID:", msg.ToUUID,
		"fromUsername:", msg.FromUsername,
		"toUsername:", msg.ToUsername,
		"fromLoc:", fmt.Sprintf("%v,%v", msg.FromLoc.Lat, msg.FromLoc.Lng),
		"toLoc:", fmt.Sprintf("%v,%v", msg.ToLoc.Lat, msg.ToLoc.Lng),
		"targetLoc:", fmt.Sprintf("%v,%v", msg.TargetLoc.Lat, msg.TargetLoc.Lng),
		"DistanceToTarget:", msg.DistanceToTarget,
		"DistanceToSender:", msg.DistanceToSender,
		"message:", msg.MsgRows,
		"radius:", msg.Radius,
		"timer:", msg.Timer,
		"createdAt:", msg.CreatedAt,
	)

	message, err := json.Marshal(msg)
	if err != nil {
		message = []byte("{\"error\": \"There was a problem handling the broadcast message.\"}")
	}

	messageBytes := []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(message)))
	_, err = toConn.writer.Write(messageBytes)
	if err != nil {
		lg.Errorf("[sendToOne] Error getting writer %s: %v", msg.ToUUID, err)
		sCtx.removeConnection(toConn)
	}
	toConn.flusher.Flush()

	return nil
}

// Send message with error back to sender
func (sCtx *serverCtx) sendError(conn *connection, msg message, errorMsg string) {
	msg.MessageType = msgError
	msg.Error = errorMsg
	message, err := json.Marshal(msg)
	if err != nil {
		message = []byte("{\"error\": \"There was a problem handling the broadcast message.\"}")
	}

	messageBytes := []byte(fmt.Sprintf("event: error\ndata: %s\n\n", string(message)))
	_, err = conn.writer.Write(messageBytes)
	if err != nil {
		lg.Errorf("[sendError] Error getting writer %s: %v", msg.ToUUID, err)
		sCtx.removeConnection(conn)
	}
	conn.flusher.Flush()
}

// --------------------------------------------------------------------------------
// If we don't have a connnection instance for this UUID, we look in the Algolia database to get the objectID,
// create a connection instance, and add to the list of connections.
// If we don't have a record in the Algolia database, we create one.

// todo: Change to Redis slowly and carefully !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

func (sCtx *serverCtx) resolveConnection(clientUuid string) *connection {
	sCtx.mutex.Lock()
	defer sCtx.mutex.Unlock()

	conn := sCtx.connections[clientUuid]
	if conn == nil {
		lg.Infof("[resolveConnection] No connection object for UUID %s. Creating a new connection.", clientUuid)
		conn = &connection{
			clientUuid: clientUuid,
		}
		sCtx.connections[clientUuid] = conn
	}

	algoliaObjectId := algolia.GetObjectIdByClientUuid(sCtx.algoliaIndex, clientUuid)
	if algoliaObjectId == "" {
		lg.Infof("[resolveConnection] No location entry in Algolia for clientUuid: %s. Creating a new Algolia entry.", clientUuid)
		conn.algoliaObjectId = algolia.Create(sCtx.algoliaIndex, clientUuid, "")
	}

	lg.Infof("[resolveConnection] for clientUuid %s (algoliaObjectId:%s, conn=nil:%v)", clientUuid, algoliaObjectId, conn == nil)

	return conn
}

// --------------------------------------------------------------------------------
func (sCtx *serverCtx) getConnection(clientUuid string) *connection {
	conn := sCtx.connections[clientUuid]
	if conn == nil {
		lg.Warnf("[getConnection]: No connection found for clientUuid: %s", clientUuid)
		return nil
	}
	return conn
}

// --------------------------------------------------------------------------------
// todo change to redis
// Remove a connection from the list of connections and from the Algolia database.
func (sCtx *serverCtx) removeConnection(conn *connection) {
	if sCtx == nil {
		lg.Warn("[removeConnection] Context is nil, so can't remove the connection.")
		return
	}

	if conn == nil {
		lg.Warn("[removeConnection] Connection is nil, so can't remove it")
		return
	}

	if conn.clientUuid == "" {
		lg.Warn("[removeConnection] Conn.clientUuid is nil, so can't remove it")
		return
	}

	sCtx.mutex.Lock()
	defer sCtx.mutex.Unlock()
	// todo Change to redis
	lg.Infof("[removeConnection] User %s w objID %s has disconnected. Removing it from local ctx and Algolia", conn.clientUuid, conn.algoliaObjectId)
	delete(sCtx.connections, conn.clientUuid)
	// change to redis
	algolia.DeleteLocation(sCtx.algoliaIndex, conn.algoliaObjectId, conn.clientUuid)
	// sCtx.redCtx.DeleteClientGeo(conn.clientUuid)
}

// --------------------------------------------------------------------------------
// Helpers

func coordAsFloat(coordStr string) float64 {
	coord, err := strconv.ParseFloat(strings.TrimSpace(coordStr), 64)
	if err != nil {
		lg.Errorf("Error converting coord to int: %v", err)
		coord = 0
	}
	return coord
}
