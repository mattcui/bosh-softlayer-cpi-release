package instance

import (
	"bytes"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

func (vg SoftlayerVirtualGuestService) SetMetadata(id int, vmMetadata Metadata) error {
	tags, err := vg.extractTagsFromVMMetadata(vmMetadata)
	if err != nil {
		return bosherr.WrapError(err, "Extracting tags from vm metadata")
	}

	err = vg.softlayerClient.SetTags(id, tags)
	if err != nil {
		return bosherr.WrapErrorf(err, "Settings tags on virtualGuest '%d'", id)
	}

	return nil
}

func (vg SoftlayerVirtualGuestService) extractTagsFromVMMetadata(vmMetadata Metadata) (string, error) {
	var tagStringBuffer bytes.Buffer
	var i int
	for key, value := range vmMetadata {
		if key == "compiling" || key == "job" || key == "index" || key == "deployment" || key == "deleted" {
			stringValue, err := value.(string)
			if !err {
				return "", bosherr.Errorf("Converting tags metadata value `%v` to string failed", value)
			}
			tagStringBuffer.WriteString(key + ":" + stringValue)
			if i != len(vmMetadata)-1 {
				tagStringBuffer.WriteString(", ")
			}
		}
		i++
	}

	return tagStringBuffer.String(), nil
}
