package dht

import (
	"fmt"

	metaffi "github.com/MetaFFI/lang-plugin-go/api"
	goruntime "github.com/MetaFFI/lang-plugin-go/go-runtime"
	"github.com/MetaFFI/plugin-sdk/compiler/go/IDL"
)

var openjdkRuntime *metaffi.MetaFFIRuntime
var chordModule *metaffi.MetaFFIModule
var newChord func(...interface{}) ([]interface{}, error)
var joinChord func(...interface{}) ([]interface{}, error)
var set func(...interface{}) ([]interface{}, error)
var get func(...interface{}) ([]interface{}, error)
var pdelete func(...interface{}) ([]interface{}, error)
var getAllKeys func(...interface{}) ([]interface{}, error)
var isFirst func(...interface{}) ([]interface{}, error)

func init() {
	var err error

	// Load OpenJDK runtime
	openjdkRuntime = metaffi.NewMetaFFIRuntime("openjdk")

	// Load Chord module
	chordModule, err = openjdkRuntime.LoadModule("./dht/Chord.class")
	if err != nil {
		fmt.Println("error chordModule")
		panic(err)
	}

	newChord, err = chordModule.Load("class=dht.Chord,callable=<init>",
		[]IDL.MetaFFIType{IDL.STRING8, IDL.INT32},
		[]IDL.MetaFFIType{IDL.HANDLE})
	if err != nil {
		fmt.Println("error newChord")
		panic(err)
	}

	// Load joinChord constructor
	joinChord, err = chordModule.Load("class=dht.Chord,callable=<init>",
		[]IDL.MetaFFIType{IDL.STRING8, IDL.STRING8, IDL.INT32},
		[]IDL.MetaFFIType{IDL.HANDLE})
	if err != nil {
		fmt.Println("error joinChord")
		panic(err)
	}
	// Load set method
	set, err = chordModule.Load("class=dht.Chord,callable=set,instance_required",
		[]IDL.MetaFFIType{IDL.HANDLE, IDL.STRING8, IDL.STRING8}, nil)
	if err != nil {
		fmt.Println("error set")
		panic(err)
	}

	// Load get method
	get, err = chordModule.Load("class=dht.Chord,callable=get,instance_required",
		[]IDL.MetaFFIType{IDL.HANDLE, IDL.STRING8},
		[]IDL.MetaFFIType{IDL.STRING8})
	if err != nil {
		fmt.Println("error get")
		panic(err)
	}

	// Load delete method
	pdelete, err = chordModule.Load("class=dht.Chord,callable=delete,instance_required",
		[]IDL.MetaFFIType{IDL.HANDLE, IDL.STRING8}, nil)
	if err != nil {
		fmt.Println("error pdelete")
		panic(err)
	}

	// Load getAllKeys method
	getAllKeys, err = chordModule.LoadWithAlias("class=dht.Chord,callable=getAllKeys,instance_required",
		[]IDL.MetaFFITypeInfo{{StringType: IDL.HANDLE}},
		[]IDL.MetaFFITypeInfo{{StringType: IDL.STRING8_ARRAY, Dimensions: 1}})
	if err != nil {
		fmt.Println("error getAllKeys")
		panic(err)
	}

	// Load isFirst method
	isFirst, err = chordModule.Load("class=dht.Chord,field=isFirst,getter,instance_required",
		[]IDL.MetaFFIType{IDL.HANDLE},
		[]IDL.MetaFFIType{IDL.BOOL})
	if err != nil {
		fmt.Println("error isFirst")
		panic(err)
	}
}

type Chord struct {
	handle goruntime.MetaFFIHandle
}

func NewChord(name string, port int32) (*Chord, error) {
	h, err := newChord(name, port)

	if err != nil {
		return nil, err
	}

	c := &Chord{}
	c.handle = h[0].(goruntime.MetaFFIHandle)
	return c, nil
}

func JoinChord(name string, rootNodeName string, port int32) (*Chord, error) {
	fmt.Printf("JoinChord called with name: %s, rootNodeName: %s, port: %d\n", name, rootNodeName, port)
	h, err := joinChord(name, rootNodeName, port)

	if err != nil {
		return nil, err
	}
	c := &Chord{}
	c.handle = h[0].(goruntime.MetaFFIHandle)
	return c, nil
}

func (c *Chord) IsFirst() (bool, error) {
	result, err := isFirst(c.handle)
	if err != nil {
		return false, err
	}
	return result[0].(bool), nil
}

func (c *Chord) Set(key string, val string) error {
	_, err := set(c.handle, key, val)
	return err
}

func (c *Chord) Get(key string) (string, error) {
	result, err := get(c.handle, key)
	if err != nil {
		return "", err
	}
	return result[0].(string), nil
}

func (c *Chord) Delete(key string) error {
	_, err := pdelete(c.handle, key)
	return err
}

func (c *Chord) GetAllKeys() ([]string, error) {
	result, err := getAllKeys(c.handle)
	if err != nil {
		return nil, err
	}
	return result[0].([]string), nil
}
