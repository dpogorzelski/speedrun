package cloud

// type Instance interface {
// 	Authorize(key []byte) error
// }

type Instance struct {
	PrivateAddress string
	PublicAddress  string
	Name           string
}

// func (i *GCEInstance) Authorize(key []byte) error {
// 	item, err := createMetadataItem(key)
// 	if err != nil {
// 		return err
// 	}

// 	hasKey, same, i := hasItem(i.Metadata, item)
// 	var items []*compute.MetadataItems

// 	if hasKey && same {
// 		return nil
// 	} else if hasKey && !same {
// 		items = updateMetadata(i.Metadata, item, i)
// 	} else if !hasKey {
// 		items = appendToMetadata(i.Metadata, item)
// 	}

// 	metadata := GCPMetadata{
// 		Fingerprint: i.Metadata.Fingerprint,
// 		Items:       items,
// 	}

// 	i.Metadata = &metadata
// 	s := strings.Split(i.Zone, "/")
// 	zone := s[len(s)-1]
// 	_, err = c.gce.Instances.SetMetadata(c.Project, zone, i.Name, i.Metadata).Do()
// 	if err != nil {
// 		return fmt.Errorf("%s failed to update metadata: ", err)
// 	}
// 	return nil
// 	i.Metadata.AddKey(key)
// }
