package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	statChannelSize = 1024
	maxFloor        = 5
)

type House struct {
	city   string
	street string
	house  string
	floor  uint
}

func (h *House) reset() {
	h.city = ""
	h.street = ""
	h.house = ""
	h.floor = 0
}

type Address struct {
	city   string
	street string
	house  string
}

type AddressMap map[Address]uint
type FloorMap map[string][]uint // key - city, count of houses with floors (from 1 to 5, 6 - reserved for overflow)

type Statistic struct {
	Items  uint
	Broken uint

	Addresses AddressMap
	Floors    FloorMap
}

func (stat *Statistic) add(h *House) {
	// check for complete item
	if len(h.city) > 0 && len(h.house) > 0 && len(h.street) > 0 && h.floor > 0 {
		stat.Items++
		address := Address{
			city:   h.city,
			street: h.street,
			house:  h.house,
		}
		if _, ok := stat.Addresses[address]; ok {
			stat.Addresses[address]++
			// first win, so don't update floors
		} else {
			stat.Addresses[address] = 1
			v, ok := stat.Floors[address.city]
			if !ok {
				v = make([]uint, maxFloor+1)
			}

			if h.floor > 5 {
				v[5]++
			} else {
				v[h.floor-1]++
			}

			if !ok {
				stat.Floors[address.city] = v
			}
		}
	}
}

func (stat *Statistic) Init() {
	stat.Items = 0
	stat.Broken = 0

	stat.Addresses = make(AddressMap)
	stat.Floors = make(FloorMap)
}

func (stat *Statistic) Merge(s *Statistic) {
	stat.Items += s.Items
	stat.Broken += s.Broken
	for address, count := range s.Addresses {
		stat.Addresses[address] += count
	}
	for city, floors := range s.Floors {
		v, ok := stat.Floors[city]
		if !ok {
			v = make([]uint, maxFloor+1)
			stat.Floors[city] = v
		}
		for i := range floors {
			v[i] += floors[i]
		}
	}
}

func (stat *Statistic) Dump() {
	fmt.Printf("Total items:  %12d\n\n", len(stat.Addresses))
	for city, floors := range stat.Floors {
		fmt.Printf("%s: houses - ", city)
		for i, floor := range floors {
			if i == 5 {
				if floor > 0 {
					fmt.Printf(", %d with floors > 5", floor)
				}
			} else {
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%d with %d floors", floor, i+1)
			}
		}
		fmt.Printf("\n")
	}

	fmt.Printf("\nTotal readed items:  %12d\n", stat.Items)
	fmt.Printf("Broken readed items: %12d\n", stat.Broken)

	for address, count := range stat.Addresses {
		if count > 1 {
			fmt.Printf("Duplicate:  %s, %s, %s\n", address.city, address.street, address.house)
		}
	}
}

type State int

const (
	waitRoot State = iota
	waitItem
	waitName
)

const (
	ItemRoot  = "<root>"
	Item      = "<item"
	ItemClose = "/>"
)

func statProcessor(pushChan <-chan House, statChan chan<- *Statistic) {
	stat := &Statistic{}
	stat.Init()
	for house := range pushChan {
		stat.add(&house)
	}
	statChan <- stat
}

type AddressParser struct {
	Stat     Statistic // don't update in parse code, push to statChan for backgroud processing
	pushChan chan House
	statChan chan *Statistic
	broken   uint

	line string
}

func (p *AddressParser) Init() {
	p.pushChan = make(chan House, statChannelSize)
	p.statChan = make(chan *Statistic)
	go statProcessor(p.pushChan, p.statChan)
}

func (p *AddressParser) resetLine(line string) {
	p.line = line
}

func (p *AddressParser) IsRoot() bool {
	if index := strings.Index(p.line, ItemRoot); index == -1 {
		return false
	} else {
		p.line = p.line[index+len(ItemRoot):]
		return true
	}
}

func (p *AddressParser) IsItem() bool {
	if index := strings.Index(p.line, Item); index == -1 {
		return false
	} else {
		p.line = p.line[index+len(Item):]
		return true
	}
}

func (p *AddressParser) IsItemClose() bool {
	p.TrimLeftSpace()
	closed := strings.HasPrefix(p.line, ItemClose)
	if closed {
		p.line = p.line[len(ItemClose):]
	}
	return closed
}

func (p *AddressParser) ExtractName() (string, bool) {
	p.TrimLeftSpace()
	if eqIdx := strings.Index(p.line, "="); eqIdx > 0 {
		name := p.line[0:eqIdx]
		p.line = p.line[eqIdx+1:]
		return name, true // don't need validate, name check will be later
	} else {
		return "", false
	}
}

func (p *AddressParser) ExtractQuotedValue() (string, bool) {
	if p.line[0] == '"' {
		if quoteEnd := strings.IndexRune(p.line[1:], '"'); quoteEnd >= 0 {
			quoteEnd += 1
			value := p.line[1:quoteEnd]
			p.line = p.line[quoteEnd+1:]
			return value, true
		}
		return p.line, false // value don't closed quote
	} else {
		return "", false
	}
}

func (p *AddressParser) TrimLeftSpace() bool {
	if p.line[0] == ' ' || p.line[0] == '\t' {
		p.line = TrimLeftAnyByte(p.line, []byte{' ', '\t'})
		//p.line = strings.TrimLeft(p.line, " \t")
		return true
	}
	return false
}

func (p *AddressParser) Empty() bool {
	return len(p.line) == 0
}

func (p *AddressParser) read(rd *bufio.Reader) error {
	var errExit error
	var house House

	var state State

	for {
		if errExit != nil {
			if errExit == io.EOF {
				errExit = nil
			}
			break
		}
		line, err := rd.ReadString('>')
		if err != nil {
			if err == io.EOF {
				errExit = err
			} else {
				errExit = err
				break
			}
		}
		p.resetLine(strings.TrimFunc(line, CutFunc))

	NEW_LINE:
		for {
			switch state {
			case waitRoot:
				if !p.IsRoot() {
					break NEW_LINE
				}
				state = waitItem
			case waitItem:
				if !p.IsItem() {
					break NEW_LINE
				}
				state = waitName
				// TODO detect if endline
				if !p.TrimLeftSpace() {
					if !p.Empty() {
						p.broken++
						house.reset()
					}
					state = waitItem
					break NEW_LINE
				}
			case waitName:
				if p.IsItemClose() {
					p.pushChan <- house
					house.reset()
					state = waitItem
					break NEW_LINE
				}
				if p.Empty() {
					break NEW_LINE
				}
				name, found := p.ExtractName()
				if found {
					switch name {
					case "city":
						if len(house.city) > 0 {
							p.broken++
							log.Printf("WARN: duplicate city in: %s\n", strings.Trim(line, "\r\n"))
							house.reset()
							state = waitItem // item is broken and skipped, wait for new item
							continue
						}
						house.city, found = p.ExtractQuotedValue()
						if !found || len(house.city) == 0 {
							p.broken++
							log.Printf("WARN: can't extract city name in: %s\n", strings.Trim(line, "\r\n"))
							house.reset()
							state = waitItem // item is broken and skipped, wait for new item
						}
					case "street":
						if len(house.street) > 0 {
							p.broken++
							log.Printf("WARN: duplicate street in: %s\n", strings.Trim(line, "\r\n"))
							house.reset()
							state = waitItem // item is broken and skipped, wait for new item
							continue
						}
						house.street, found = p.ExtractQuotedValue()
						if !found || len(house.street) == 0 {
							p.broken++
							log.Printf("WARN: can't extract street name in: %s\n", strings.Trim(line, "\r\n"))
							house.reset()
							state = waitItem // item is broken and skipped, wait for new item
						}
					case "house":
						if len(house.house) > 0 {
							p.broken++
							log.Printf("WARN: duplicate house in: %s\n", strings.Trim(line, "\r\n"))
							house.reset()
							state = waitItem // item is broken and skipped, wait for new item
							continue
						}
						house.house, found = p.ExtractQuotedValue()
						if !found || len(house.street) == 0 {
							p.broken++
							log.Printf("WARN: can't extract house name in: %s\n", strings.Trim(line, "\r\n"))
							house.reset()
							state = waitItem // item is broken and skipped, wait for new item
						}
					case "floor":
						var floor string
						if house.floor > 0 {
							p.broken++
							log.Printf("WARN: duplicate floor in: %s\n", strings.Trim(line, "\r\n"))
							house.reset()
							state = waitItem // item is broken and skipped, wait for new item
							continue
						}
						floor, found = p.ExtractQuotedValue()
						if found {
							var err error
							var fl uint64
							fl, err = strconv.ParseUint(floor, 10, 16)
							if err != nil || fl == 0 {
								log.Printf("WARN: incorrect house floor in: %s\n", strings.Trim(line, "\r\n"))
								house.reset()
								state = waitItem // item is broken and skipped, wait for new item
							} else {
								house.floor = uint(fl)
							}
						} else if !found {
							p.broken++
							log.Printf("WARN: can't extract house floor in: %s\n", strings.Trim(line, "\r\n"))
							house.reset()
							state = waitItem // item is broken and skipped, wait for new item
						}
					default:
						if _, found = p.ExtractQuotedValue(); !found {
							p.broken++
							log.Printf("WARN: can't extract %s value in: %s\n", name, strings.Trim(line, "\r\n"))
							house.reset()
							state = waitItem // item is broken and skipped, wait for new item
						}
					}
				} else {
					p.broken++
					log.Printf("WARN: incomplete line: %s\n", strings.Trim(line, "\r\n"))
					state = waitItem // item is broken and skipped, wait for new item
					continue
				}
			}
		}
	}

	if errExit == nil {
		p.pushChan <- house
	}

	close(p.pushChan)
	stat := <-p.statChan
	p.Stat = *stat
	p.Stat.Broken += p.broken // merge broken stat
	p.broken = 0

	return errExit
}

func (p *AddressParser) readAddressFile(filePath string) error {
	f, err := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	bufRd := bufio.NewReader(f)

	return p.read(bufRd)
}
