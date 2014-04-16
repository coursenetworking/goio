package goio

type User struct {
	Event
	Id      string
	Clients *Clients
	Rooms   *Rooms
}

func (self *User) Receive(message *Message) {
	self.Clients.Receive(message)
}

func (self *User) Has(id string) bool {
	return nil != self.Clients.Get(id)
}

func (self *User) Delete(id string) {
	self.Clients.Delete(id)
	if 0 == self.Clients.Count() {
		self.Destory()
	}
}

func (self *User) Destory() {
	self.Emit("destory", &Message{
		EventName: "destory",
		CallerId:  self.Id,
	})
}

func (self *User) Add(clt *Client) {
	if self.Has(clt.Id) {
		return
	}

	clt.User = self
	clt.On("destory", func(message *Message) {
		clt.User.Delete(clt.Id)
	})

	// we add it to global clients too
	GlobalClients().Add(clt)
	self.Clients.Add(clt)
}
