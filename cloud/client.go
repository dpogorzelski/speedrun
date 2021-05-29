package cloud

type Client struct {
	gcp *gcpClient
}

func NewClient(options ...ConfigOption) (*Client, error) {
	var config settings

	for _, opt := range options {
		opt.Set(&config)
	}

	gclient, err := newGCPClient(config.gcp)
	if err != nil {
		return nil, err
	}
	c := &Client{
		gcp: gclient,
	}

	return c, nil
}

func (c *Client) GetInstances(filter string) ([]Instance, error) {
	instances := []Instance{}
	gcpInstances, err := c.gcp.GetInstances(filter)
	if err != nil {
		return instances, err
	}

	for _, instance := range gcpInstances {
		i := &Instance{
			PrivateAddress: instance.NetworkInterfaces[0].NetworkIP,
			PublicAddress:  instance.NetworkInterfaces[0].AccessConfigs[0].NatIP,
			Name:           instance.Name,
		}
		instances = append(instances, *i)
	}

	return instances, err
}

func (c *Client) AuthorizeKey(authorizedKey []byte) error {
	err := c.gcp.addKeyToMetadata(authorizedKey)
	if err != nil {
		return err
	}

	err = c.gcp.addUserKey(authorizedKey)
	return err
}

func (c *Client) RevokeKey(authorizedKey []byte) error {
	err := c.gcp.removeUserKey(authorizedKey)
	if err != nil {
		return err
	}

	err = c.gcp.removeKeyFromMetadata(authorizedKey)
	if err != nil {
		return err
	}

	return err
}

func (c *Client) ListKeys() error {
	err := c.gcp.listUserKeys()
	if err != nil {
		return err
	}
	return err
}

// func (c *Client) AuthorizeKeyInstance(authorizedKey []byte, instancePool *InstancePool) error {
// 	pool := pond.New(10, 0, pond.MinWorkers(10))
// 	for _, instance := range instancePool.instances {
// 		pool.Submit(func() {
// 			instance.Authorize(authorizedKey)
// 		})
// 	}
// 	pool.StopAndWait()

// 	return nil
// }
