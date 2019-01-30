package main

type Account struct {
	id            uint32
	email         string
	fName         string
	sName         string
	phone         string //optimise uint64
	sex           bool
	birth         uint32 //optimise Ограничено снизу 01.01.1950 и сверху 01.01.2005-ым
	country       string
	city          string
	joined        uint32 //optimise снизу 01.01.2011, сверху 01.01.2018
	status        int8   //0 - free, 1 - complicated, 2 - busy
	interests     []string
	premiumStart  uint32
	premiumFinish uint32
	likeIds       []uint32
	likeTss       []uint32
}

type Indexes struct {
	accounts map[uint32]struct{}
	emails   map[string]struct{}
	phones   map[string]struct{}
}
