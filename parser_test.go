package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddressParser(t *testing.T) {
	xml := strings.NewReader(
		`<?xml version="1.0" encoding="utf-8"?>
<root>
<item city="Барнаул" street="Дальняя улица" house="56" floor="2" />

<item />

<item house="6" floor="3" />
<item city= street="Дальняя улица" house="6" floor="3" />

<item city="Братск" street="Большая Октябрьская улица" house="65" floor="5" />
<item city="Барнаул" street="Дальняя улица" house="6" floor="3" />
<item city="Барнаул" street="Ближняя улица" house="6" floor="2" />
<item city="Барнаул" street="Дальняя улица" house="56" floor="3" />

<item city="Братск" street="Большая Октябрьская улица" house="66" floor="7" />
</root>
`)

	wantStat := Statistic{
		Items:  6,
		Broken: 2,
		Addresses: AddressMap{
			{
				city:   "Барнаул",
				street: "Дальняя улица",
				house:  "56",
			}: 2,
			{
				city:   "Барнаул",
				street: "Дальняя улица",
				house:  "6",
			}: 1,
			{
				city:   "Барнаул",
				street: "Ближняя улица",
				house:  "6",
			}: 1,
			{
				city:   "Братск",
				street: "Большая Октябрьская улица",
				house:  "65",
			}: 1,
			{
				city:   "Братск",
				street: "Большая Октябрьская улица",
				house:  "66",
			}: 1,
		},
		Floors: FloorMap{
			"Барнаул": []uint{0, 2, 1, 0, 0, 0},
			"Братск":  []uint{0, 0, 0, 0, 1, 1},
		},
	}

	var p AddressParser
	p.Init()

	err := p.read(bufio.NewReader(xml))
	require.NoErrorf(t, err, "xml read error")

	assert.Equalf(t, wantStat.Items, p.Stat.Items, "items")
	assert.Equalf(t, wantStat.Broken, p.Stat.Broken, "broken")

	for address, count := range wantStat.Addresses {
		if realCount, ok := p.Stat.Addresses[address]; ok {
			if realCount != count {
				t.Errorf(" - Address %+v: %d", address, count)
				t.Errorf(" + Address %+v: %d", address, realCount)
			}
		} else {
			t.Errorf(" - Address %+v: %d", address, count)
		}
	}
	for address, realCount := range p.Stat.Addresses {
		if _, ok := wantStat.Addresses[address]; !ok {
			t.Errorf(" + Address %+v: %d", address, realCount)
		}
	}

	for city, floors := range wantStat.Floors {
		if realFloors, ok := p.Stat.Floors[city]; ok {
			if !reflect.DeepEqual(realFloors, floors) {
				t.Errorf(" - %s Floors %+v", city, floors)
				t.Errorf(" + %s Floors %+v", city, realFloors)
			}
		} else {
			t.Errorf(" - %s Floors %+v", city, floors)
		}
	}
	for city, realFloors := range p.Stat.Floors {
		if _, ok := wantStat.Floors[city]; !ok {
			t.Errorf(" + %s Floors %+v", city, realFloors)
		}
	}
}

func BenchmarkAddressParserSmall(b *testing.B) {
	xml :=
		`<?xml version="1.0" encoding="utf-8"?>
<root>
<item city="Барнаул" street="Дальняя улица" house="56" floor="2" />
<item city="Братск" street="Большая Октябрьская улица" house="65" floor="5" />
<item city="Балаково" street="Барыши, местечко" house="67" floor="2" />
<item city="Азов" street="Просека, улица" house="156" floor="3" />
<item city="Видное" street="Авиаторов, улица" house="185" floor="3" />
<item city="Братск" street="7-я Вишнёвая улица" house="49" floor="5" />
<item city="Батайск" street="Мостотреста, улица" house="133" floor="4" />
<item city="Великий Новгород" street="Филимонковская улица" house="44" floor="1" />
<item city="Абакан" street="2-я Валуевская улица" house="172" floor="3" />
<item city="Бугульма" street="Варшавское шоссе" house="92" floor="2" />
<item city="Ачинск" street="Варшавское шоссе" house="39" floor="4" />
<item city="Балашов" street="Учительский переулок" house="184" floor="1" />
<item city="Бийск" street="Сиреневая улица" house="14" floor="3" />
<item city="Армавир" street="Песчаная улица" house="69" floor="1" />
<item city="Ачинск" street="3-я Барышевская улица" house="154" floor="3" />
<item city="Астрахань" street="4-й Заречный переулок" house="100" floor="5" />
<item city="Владивосток" street="Дальняя улица" house="196" floor="3" />
<item city="Белово" street="Симферопольская улица" house="116" floor="1" />
<item city="Владивосток" street="Родниковая улица" house="105" floor="1" />
<item city="Бугульма" street="Маршала Кутахова, улица" house="91" floor="1" />
<item city="Архангельск" street="Лесхозная улица" house="85" floor="3" />
<item city="Армавир" street="Евгения Родионова, улица" house="22" floor="4" />
<item city="Белебей" street="Маршала Кутахова, улица" house="93" floor="1" />
<item city="Астрахань" street="Еловая улица" house="52" floor="5" />
<item city="Бийск" street="Вишнёвая улица" house="128" floor="5" />
<item city="Абакан" street="Симферопольская улица" house="177" floor="2" />
<item city="Благовещенск" street="Михаила Кондакова, улица" house="31" floor="2" />
<item city="Братск" street="2-я Катерная улица" house="60" floor="4" />
<item city="Буйнакск" street="Московская улица" house="142" floor="2" />
<item city="Балаково" street="Лагерная улица" house="122" floor="4" />
<item city="Абакан" street="Барышевская улица" house="69" floor="3" />
<item city="Балашов" street="1-й Богородский переулок" house="7" floor="1" />
<item city="Бердск" street="4-я Майская улица" house="53" floor="3" />
<item city="Бугульма" street="1-я Мичуринская улица" house="13" floor="4" />
<item city="Ангарск" street="5-я Майская улица" house="197" floor="3" />
<item city="Анжеро-Судженск" street="Люблинская улица" house="42" floor="1" />
<item city="Балашиха" street="Живописная улица" house="113" floor="2" />
<item city="Брянск" street="1-я Майская улица" house="20" floor="1" />
<item city="Бор" street="Барышевская Роща, улица" house="52" floor="2" />
<item city="Благовещенск" street="Спортивная улица" house="3" floor="3" />
<item city="Видное" street="Первомайская улица" house="92" floor="2" />
<item city="Астрахань" street="2-я Мичуринская улица" house="2" floor="4" />
<item city="Верхняя Пышма" street="Бутовское Кольцо, улица" house="178" floor="2" />
<item city="Арзамас" street="Герцена, улица" house="83" floor="5" />
<item city="Благовещенск" street="Центральная улица" house="161" floor="5" />
<item city="Бор" street="Большая Октябрьская улица" house="140" floor="3" />
<item city="Березовский" street="1-я Валуевская улица" house="77" floor="1" />
<item city="Артем" street="Михаила Кондакова, улица" house="177" floor="5" />
<item city="Архангельск" street="Фруктовая улица" house="135" floor="2" />
<item city="Великие Луки" street="Звёздная улица" house="188" floor="1" />
<item city="Благовещенск" street="1-й Богородский переулок" house="79" floor="1" />
<item city="Бийск" street="Вишнёвая улица" house="178" floor="3" />
<item city="Ангарск" street="Космонавтов, улица" house="123" floor="3" />
<item city="Бийск" street="Комсомольская улица" house="3" floor="2" />
<item city="Березники" street="Парковая улица" house="149" floor="1" />
<item city="Белорецк" street="Овражная улица" house="97" floor="3" />
<item city="Алексин" street="Речная улица" house="185" floor="1" />
</root>
`

	for n := 0; n < b.N; n++ {
		var p AddressParser
		xmlReader := strings.NewReader(xml)
		p.Init()
		err := p.read(bufio.NewReader(xmlReader))
		require.NoErrorf(b, err, "xml read error")
	}
}

func BenchmarkAddressParserLarge(b *testing.B) {
	_, filename, _, _ := runtime.Caller(0)
	path := path.Join(path.Dir(filename), "bench", "large.xml")
	xml, err := ioutil.ReadFile(path)
	require.NoErrorf(b, err, "xml file load error")

	for n := 0; n < b.N; n++ {
		var p AddressParser
		p.Init()
		err := p.read(bufio.NewReader(bytes.NewReader(xml)))
		require.NoErrorf(b, err, "xml read error")
	}
}
