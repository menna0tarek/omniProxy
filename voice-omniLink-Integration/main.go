package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"example.com/m/domain"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"example.com/m/callstorage"
	"example.com/m/omnilink"
)

var (
	// serverURL      = envString("OMNILINK_SERVER_ADD", "https://omnilink-api.apps.omnilink.k8sctl.qc01.cequens.local:8443/v1/events/trigger") // will be from env var as const
	serverURL      = envString("OMNILINK_SERVER_ADD", "http://localhost:8088/omnilink") // will be from env var as const
	pAddress       = envString("HTTP_ADDRESS", ":8089")
	apiKey         = envString("OMNILINK_KEY", "a05a346016cd93fffe620a8f51e221a4")                    // will be from env var a
	cbURL          = envString("CALLBACK_URL", "https://45d0-156-197-35-174.ngrok-free.app/callback") // will be from env var as co
	basicURL       = envString("BASIC_REDIRECTED_URL", "https://45d0-156-197-35-174.ngrok-free.app/calls")
	musicOnHoldXML = envString("MUSIC_ONHOLD_XML", "<begin><play url='https://www.asterisksounds.org/sites/asterisksounds.org/files/sounds/en/core/basic-pbx-ivr-main.ogg'></play></close></begin>")
	ttl            = envString("TTL_CALL_STORAGE", "10")
	callbackURL    string
)

type callStorageManager interface {
	Store(callId string, mobile string, subId string) (*callstorage.Call, error)
	Get(callId string) (*callstorage.Call, error)
	Remove(callId string) error
}

var callManager callStorageManager

func main() {
	var proxyAddress string
	proxyAddress = os.Getenv("PROXY_ADDRESS")
	if proxyAddress == "" {
		log.Println("[INFO][main] Env variable PROXY_ADDRESS not set will use default : ", pAddress)
		proxyAddress = pAddress
	}
	log.Println("[INFO][main]starting proxy listening on port : ", proxyAddress)
	callbackURL = os.Getenv("CALLBACK_URL")
	if callbackURL == "" {
		fmt.Println("[INFO][main] Env variable CALLBACK_URL not set will use default : ", cbURL)
		callbackURL = cbURL
	}
	log.Println("[TRACE][main] callback url : ", callbackURL)

	//
	errorChannel := make(chan error)
	// Init call storage manager.
	ttlnum, _ := strconv.Atoi(ttl)
	fmt.Println(ttlnum)
	callManager = callstorage.New(ttlnum)

	// Handle incoming requests
	http.Handle("/calls", chimiddleware.Recoverer(RequestLogger()(handleRequest())))

	// Handle callbacks
	http.Handle("/callback", chimiddleware.Recoverer(RequestLogger()(handleCallback())))

	// Start server
	httpserver := &http.Server{
		Addr: pAddress,
	}
	go func() {
		errorChannel <- httpserver.ListenAndServe()
	}()

	// Capture interrupts, to handle them gracefully.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errorChannel <- fmt.Errorf("got terminating signal: %s", <-c)
	}()

	// Wait for errors on the error channel, this will stall until an error is received.
	if err := <-errorChannel; err != nil {
		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownRelease()
		if err := httpserver.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("HTTP shutdown error: %v", err)
		}
		log.Fatal(err)
	}
}

func handleRequest() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("[INFO][handleRequest] New incoming request")
		if r.Method != http.MethodGet {
			http.Error(w, "[handleRequest] Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		// body, err := io.ReadAll(r.Body)
		// if err != nil {
		// 	log.Println("[ERROR][handleRequest] Error reading request body : ", err)
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		// defer r.Body.Close()

		var requestData domain.RequestData

		callId := r.URL.Query().Get("callID")
		if callId == "" {
			log.Println("[ERROR][handleRequest] callId is required")
			http.Error(w, "callId is required", http.StatusBadRequest)
			return
		}
		name := r.URL.Query().Get("name")
		// if name == "" {
		// 	log.Println("[ERROR][handleRequest] appName is required")
		// 	http.Error(w, "appName is required", http.StatusBadRequest)
		// 	return
		// }
		phone := r.URL.Query().Get("phone")

		requestData.Name = name
		requestData.CallId = callId
		requestData.To.Phone = phone
		requestData.To.SubscriberId = phone

		// if err := json.Unmarshal(body, &requestData); err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		log.Println("[INFO][handleRequest] Request AppName and callId: ", requestData.Name, requestData.CallId)

		// if requestData.To.Phone == "" {
		// 	http.Error(w, "mobile is required", http.StatusBadRequest)
		// 	return
		// }

		// if requestData.To.SubscriberId == "" {
		// 	http.Error(w, "subscriberId is required", http.StatusBadRequest)
		// 	return
		// }
		originalRedirectedUrl := serverURL
		modifiedRedirectedUrl := basicURL

		call, err := callManager.Get(requestData.CallId)
		if err != nil {
			http.Error(w, "can't store the call", http.StatusInternalServerError)
			return
		}

		if call == nil {
			fmt.Println("can't find the call")
			call, err = callManager.Store(requestData.CallId, requestData.To.Phone, requestData.To.SubscriberId)
			if err != nil {
				http.Error(w, "can't store the call", http.StatusInternalServerError)
				return
			}
		} else {
			fmt.Println("found the call")
			token := r.URL.Query().Get("token")
			fmt.Println("token: ", token)
			if token != "" {
				modifiedRedirectedUrl += "?" + token
				url := call.GetUrlByToken(token)
				fmt.Println("token: ", url)
				if url != "" {
					originalRedirectedUrl = url
				}
			}

		}

		start_time := time.Now()

		omniLinkRequestBody := &domain.OmnilinkRequestData{
			Name:          "testvoice", //TODO requestData.Name,
			TransactionId: requestData.CallId,
			To: domain.ToData{
				SubscriberId: requestData.To.SubscriberId,
				Phone:        requestData.To.Phone,
			},
			Overrides: map[string]map[string]string{
				"voice": {
					"voiceCallbackUrl": callbackURL,
				},
			},
			Payload: map[string]interface{}{},
		}
		forwardedRequestBody, err := json.Marshal(omniLinkRequestBody)
		fmt.Println("forwardedRequestBody: ", string(forwardedRequestBody))
		if err != nil {
			http.Error(w, "Failed to marshal the json", http.StatusInternalServerError)
			return
		}

		resp, err := omnilink.Request(originalRedirectedUrl, forwardedRequestBody, apiKey)
		if err != nil {
			log.Println("omnilink error : ", err)
			http.Error(w, "Error forwarding request", http.StatusInternalServerError)
		}
		log.Println("[DEBUG][handleRequest] Response from omnilink : ", string(resp))
		log.Println("[Trace][handleRequest] Waiting for the callback")

		select {
		case result, ok := <-call.InboundChannel:
			if !ok {
				http.Error(w, "Error while waiting to  receive the callback ", http.StatusInternalServerError)
				return
			}
			log.Println("[INFO][handleRequest] Callback received after : ", time.Since(start_time))
			log.Println("[INFO][handleRequest] Callback received with answer : ", result)

			w.WriteHeader(http.StatusOK)
			if result.RedirectURL == "" {
				fmt.Println("Lets remove call no more redirection .")
				callManager.Remove(requestData.CallId) // Remove this call as it is not redirecting anymore.
			}
			// response, _ := json.Marshal(result.ExecutionPlan)
			log.Println("***** ****** ***** [INFO][handleRequest] Response to the server : ", result.ExecutionPlan)
			w.Write([]byte(result.ExecutionPlan))
			return

		case <-time.After(10 * time.Second):
			log.Printf("[ERROR][handleRequest] Timeout waiting for callback for %s \n", requestData.CallId)

			waitingData := domain.CallbackRequestData{
				To:                     requestData.To.Phone,
				CallID:                 requestData.CallId,
				StatusCallbackEndpoint: "",
				RedirectURL:            modifiedRedirectedUrl,
				RingingDuration:        10,
				ExecutionPlan:          musicOnHoldXML,
			}
			strwaitingData, _ := json.Marshal(waitingData)
			log.Println(strwaitingData)
			w.Write([]byte(musicOnHoldXML))
			// http.Error(w, "Timeout waiting for callback", http.StatusRequestTimeout)
		}
	})
}

func handleCallback() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("[TRACE][handleCallback]  New incoming CALLBACK request")
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read callback body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("[ERROR][handleCallback] Error reading request body : ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("[handleCallback] Callback body : ", string(body))
		defer r.Body.Close()

		// Extract callID
		data := &domain.CallbackRequestData{}
		if err := json.Unmarshal(body, &data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("[handleCallback] Callback data as map : ")
		if data.CallID == "" {
			log.Println("[ERROR][handleCallback] callid is required")
			http.Error(w, "callid is required", http.StatusBadRequest)
			return
		}
		log.Println("[INFO][handleCallback] Callback response for Callid : ", data)

		//response to the server using w
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Callback received"))

		call, err := callManager.Get(data.CallID)
		if err != nil {
			fmt.Println("Error with getting the call : ", err)
		}
		if data.RedirectURL != "" {
			token := call.ReplaceUrl(data.RedirectURL)
			data.RedirectURL = basicURL + "?token=" + token + "&callID=" + data.CallID + "&phone=" + data.To
		}

		forwardedRequestBody, _ := json.Marshal(data)
		modifiedXML := omnilink.Extract(forwardedRequestBody)
		data.ExecutionPlan = modifiedXML
		call.InboundChannel <- *data
	})
}

func envString(key string, fallback string) string {
	if value, ok := syscall.Getenv(key); ok {
		return value
	}
	return fallback
}

// RequestLogger will log the incoming request and its response.
func RequestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			buf := new(bytes.Buffer)
			rw.Tee(buf)
			entry := &logEntry{
				ctx:      r.Context(),
				endpoint: r.URL.Path,
			}
			start := time.Now()

			// Defer the write on log entry to catch status code, time elapsed and response body.
			defer func() {
				var body []byte
				if rw.Status() < 500 {
					body, _ = io.ReadAll(buf)
				}
				log.Printf(
					"%s %s %s %d %d %s",
					r.Method,
					r.URL.Path,
					r.Proto,
					rw.Status(),
					time.Since(start), string(body))
			}()
			next.ServeHTTP(rw, chimiddleware.WithLogEntry(r, entry))
		})
	}
}

// logEntry implements the chi LogEntry interface.
type logEntry struct {
	ctx      context.Context
	endpoint string
}

func (l *logEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {

	fields := []interface{}{
		"endpoint", l.endpoint,
		"status", status,
		"bytes", bytes,
		"elapsed", elapsed,
	}

	if body, ok := extra.([]byte); ok {
		if len(body) > 0 {
			fields = append(fields, "body", string(body))
		}
	}

	fmt.Println("http.request", fields)
}

func (l *logEntry) Panic(v interface{}, stack []byte) {
	fmt.Println("[PANIC]",
		"stack_trace", stack,
	)
}
