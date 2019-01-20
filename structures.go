package main

type Account struct {
	id        uint32
	email     string
	fName     string
	sName     string
	phone     string //optimise uint64
	sex       bool
	birth     uint32 //optimise Ограничено снизу 01.01.1950 и сверху 01.01.2005-ым
	country   string
	city      string
	joined    uint32 //optimise снизу 01.01.2011, сверху 01.01.2018
	status    int8   //0 - free, 1 - complicated, 2 - busy
	interests *[]string
	premium   Premium
	likes     *[]Like
}

type Premium struct {
	start  uint32 //optimise с нижней границей 01.01.2018
	finish uint32
}

type Like struct {
	id uint32
	ts uint32
}

type Schema struct {
	accounts map[uint32]*Account
	emails   map[string]struct{}
	phones   map[string]struct{}
}
