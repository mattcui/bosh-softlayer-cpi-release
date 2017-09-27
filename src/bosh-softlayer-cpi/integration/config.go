package integration

import (
	"bosh-softlayer-cpi/action"
	boshapi "bosh-softlayer-cpi/api"
	boshdisp "bosh-softlayer-cpi/api/dispatcher"
	"bosh-softlayer-cpi/api/transport"
	boshcfg "bosh-softlayer-cpi/config"
	"bosh-softlayer-cpi/softlayer/client"
	vpsVm "bosh-softlayer-cpi/softlayer/vps_service/client/vm"
	"bytes"
	"encoding/json"
	"fmt"
	boshlogger "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/softlayer/softlayer-go/session"
	"io"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"strings"
)

var (
	// A stemcell that will be created/loaded in integration_suite_test.go
	existingStemcellId         string
	in, out, errOut, errOutLog bytes.Buffer
	username                   = envRequired("SL_USERNAME")
	apiKey                     = envRequired("SL_API_KEY")

	// Configurable defaults
	stemcellId   = envOrDefault("STEMCELL_ID", "1633205")
	stemcellFile = envOrDefault("STEMCELL_FILE", "")
	stemcellUuid = envOrDefault("DATACENTER", "ea065435-f7ec-4f1c-8f3f-2987086b1427")
	datacenter   = envOrDefault("DATACENTER", "lon02")
	ipAddrs      = strings.Split(envOrDefault("PRIVATE_IP", "192.168.100.102,192.168.100.103,192.168.100.104"), ",")

	ts           *httptest.Server
	cfg          boshcfg.Config
	boshResponse boshdisp.Response
	vps          *vpsVm.Client

	// Channel that will be used to retrieve IPs to use
	ips chan string

	trace      = false
	timeout    = 50000
	cfgContent = fmt.Sprintf(
		`{
		  "cloud": {
		    "plugin": "softlayer",
		    "properties": {
		      "softlayer": {
			    "username": "%s",
			    "api_key": "%s"
		      },
		      "registry": {
			    "user": "registry",
			    "password": "1330c82d-4bc4-4544-4a90-c2c78fa66431",
			    "address": "127.0.0.1",
			    "http": {
			      "port": 8000,
			      "user": "registry",
			      "password": "1330c82d-4bc4-4544-4a90-c2c78fa66431"
			    },
			    "endpoint": "http://registry:1330c82d-4bc4-4544-4a90-c2c78fa66431@127.0.0.1:8000"
		      },
		      "agent": {
			    "ntp": [
			    ],
			    "blobstore": {
			      "provider": "dav",
			      "options": {
			        "endpoint": "http://127.0.0.1:25250",
			        "user": "agent",
			        "password": "agent"
			      }
			    },
			    "mbus": "nats://nats:nats@127.0.0.1:4222"
		      }
		    }
		  }
		}`, username, apiKey)

	// Stuff of softlayer client
	multiWriter = io.MultiWriter(&errOut, &errOutLog)
	logger      = boshlogger.NewWriterLogger(boshlogger.LevelDebug, multiWriter, multiWriter)
	multiLogger = boshapi.MultiLogger{Logger: logger, LogBuff: &errOutLog}
	uuidGen     = uuid.NewGenerator()
	sess        *session.Session
)

func execCPI(request string) (boshdisp.Response, error) {
	var err error
	var softlayerClient client.Client

	softlayerClient = client.NewSoftLayerClientManager(sess, vps)
	actionFactory := action.NewConcreteFactory(
		softlayerClient,
		uuidGen,
		cfg,
		multiLogger,
	)

	caller := boshdisp.NewJSONCaller()
	dispatcher := boshdisp.NewJSON(actionFactory, caller, multiLogger)

	in.WriteString(request)
	cli := transport.NewCLI(&in, &out, dispatcher, multiLogger)

	var response []byte

	if err = cli.ServeOnce(); err != nil {
		return boshResponse, err
	}

	if response, err = ioutil.ReadAll(&out); err != nil {
		return boshResponse, err
	}

	if err = json.Unmarshal(response, &boshResponse); err != nil {
		return boshResponse, err
	}
	return boshResponse, nil
}

func envRequired(key string) (val string) {
	if val = os.Getenv(key); val == "" {
		panic(fmt.Sprintf("Could not find required environment variable '%s'", key))
	}
	return
}

func envOrDefault(key, defaultVal string) (val string) {
	if val = os.Getenv(key); val == "" {
		val = defaultVal
	}
	return
}