package database

func (c *Connection) AddressesTableName() string {
	return c.Chain + "_addresses"
}

func (c *Connection) AppearancesTableName() string {
	return c.Chain + "_appearances"
}

func (c *Connection) ChunksTableName() string {
	return c.Chain + "_chunks"
}
