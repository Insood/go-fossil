package main

type System interface {
	Initialize(*Game)
	Update(*Game)
}
