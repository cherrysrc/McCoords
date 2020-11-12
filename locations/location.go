package locations

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

type Location struct {
	X, Y, Z float64
}

type LocationMap map[string]Location

func New() LocationMap {
	return make(map[string]Location, 0)
}

func (lmap LocationMap) Set(name string, location Location) {
	lmap[name] = location
}

func (lmap LocationMap) Delete(name string) error {
	_, ok := lmap[name]
	if ok {
		delete(lmap, name)
	} else {
		return errors.New("Location not found!")
	}
	return nil
}

func (lmap LocationMap) ToString() string {
	out := ""
	for name, location := range lmap {
		out += fmt.Sprintf("%10s: %7.2f, %7.2f, %7.2f\n", name, location.X, location.Y, location.Z)
	}
	return out
}

func (lmap LocationMap) Save(fileName string) error {
	data, err := json.Marshal(lmap)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, data, 0644)
}

func Load(fileName string) (LocationMap, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var lmap LocationMap
	err = json.Unmarshal(data, &lmap)
	if err != nil {
		return nil, err
	}
	return lmap, nil
}
