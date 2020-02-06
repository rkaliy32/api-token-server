package slowpoke

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/rkaliy32/api-token-server/pudge"
)

func logg(i interface{}) {

	//t := time.Now()
	//fmt.Printf("%02d.%02d.%04d %02d:%02d:%02d\t%v\n",
	//	t.Day(), t.Month(), t.Year(),
	//	t.Hour(), t.Minute(), t.Second(), i)
}

func ch(err error, t *testing.T) {
	if err != nil {
		t.Error(err)
	}
}

func TestBase(t *testing.T) {
	var err error
	var b []byte
	f := "test/TestBase.db"
	DeleteFile(f)
	key := []byte("1")
	err = Set(f, key, key)
	ch(err, t)
	b, err = Get(f, key)
	if !bytes.Equal(b, key) {
		t.Error("not equal")
	}
	err = Set(f, key, []byte("3"))
	ch(err, t)
	Close(f)

	b, err = Get(f, key)
	if !bytes.Equal(b, []byte("3")) {
		t.Error("not equal")
	}
	key2 := []byte("2")
	err = Set(f, key2, key2)
	ch(err, t)
	b, err = Get(f, key2)
	if !bytes.Equal(b, key2) {
		t.Error("not equal")
	}

	_, err = Delete(f, key2)

	b, err = Get(f, key2)
	if err == nil || !bytes.Equal(b, nil) {
		t.Error("not deleted")
	}
	Close(f)

	b, err = Get(f, key2)
	if err == nil || !bytes.Equal(b, nil) {
		t.Error("not deleted")
	}
	keys, err := Keys(f, nil, 0, 0, true)
	ch(err, t)
	if !bytes.Equal(key, keys[0]) {
		t.Error("not equal")
	}
	CloseAll()
}

func TestOpen(t *testing.T) {
	d, _ := Open("test/open.db")
	//fmt.Println(d)
	Set("test/open.db", []byte("foo"), []byte("bar"))
	//val, ok := d.ReadKey("foo")
	ret, _ := Get("test/open.db", []byte("foo"))
	if bytes.Compare(ret, []byte("bar")) != 0 {
		t.Error("not bar", ret)
	}
	d.Delete("foo")
	var v string
	err := d.Get("foo", &v)
	if err != pudge.ErrKeyNotFound {
		t.Error(err)
	}

}

func TestAsync(t *testing.T) {
	len := 5
	file := "test/async.db"
	//DeleteFile(file)
	defer CloseAll()

	messages := make(chan int)
	readmessages := make(chan string)
	var wg sync.WaitGroup

	append := func(i int) {
		defer wg.Done()
		k := ("Key:" + strconv.Itoa(i))
		v := ("Val:" + strconv.Itoa(i))
		err := Set(file, []byte(k), []byte(v))
		if err != nil {
			t.Error(err)
		}
		messages <- i
	}

	read := func(i int) {
		defer wg.Done()
		k := ("Key:" + strconv.Itoa(i))
		v := ("Val:" + strconv.Itoa(i))

		b, _ := Get(file, []byte(k))

		if string(b) != string(v) {
			t.Error("not mutch")
		}
		readmessages <- fmt.Sprintf("read N:%d  content:%s", i, string(b))
	}

	for i := 1; i <= len; i++ {
		wg.Add(1)
		go append(i)

	}

	go func() {
		for i := range messages {
			_ = i
			//fmt.Println(i)
		}
	}()

	go func() {
		for i := range readmessages {
			_ = i
			//fmt.Println(i)
		}
	}()

	wg.Wait()

	for i := 1; i <= len; i++ {

		wg.Add(1)
		go read(i)
	}
	wg.Wait()

}

func TestBytesConvert(t *testing.T) {
	file := "test/BytesConvert.db"
	DeleteFile(file)
	defer CloseAll()
	for i := 1; i <= 20; i++ {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(i))
		Set(file, b, b)
		_, ee := Get(file, b)
		if ee != nil {
			t.Error(ee)
		}

	}
	b20 := make([]byte, 4)
	binary.BigEndian.PutUint32(b20, uint32(20))
	keys, _ := Keys(file, nil, 1, 0, false)
	if len(keys) != 1 || !bytes.Equal(b20, keys[0]) {
		t.Error(file)
	}
}

func TestBench(t *testing.T) {
	file := "test/bench.db"
	err := DeleteFile(file)
	if err != nil {
		fmt.Println(err)
	}
	var wg sync.WaitGroup

	appendd := func(i int) {
		defer wg.Done()
		k := []byte(fmt.Sprintf("%04d", i))
		err := Set(file, k, k)
		if err != nil {
			fmt.Println(err)
		}
	}

	t1 := time.Now()
	for i := 0; i < 100; i++ {
		wg.Add(1)
		appendd(i)
	}
	wg.Wait()
	t2 := time.Now()

	fmt.Printf("The 100 Set took %v to run.\n", t2.Sub(t1))

	read := func(i int) {
		defer wg.Done()
		k := []byte(fmt.Sprintf("%04d", i))
		_, _ = Get(file, k)
		//fmt.Println(string(res))

	}
	//_ = read
	t3 := time.Now()
	for i := 0; i < 100; i++ {
		wg.Add(1)
		read(i)
		//k := []byte(fmt.Sprintf("%04d", i))
		//_, _ = Get(file, k)
	}
	wg.Wait()
	t4 := time.Now()

	fmt.Printf("The 100 Get took %v to run.\n", t4.Sub(t3))

	//Sets
	var pairs [][]byte
	for i := 0; i < 100; i++ {
		k := []byte(fmt.Sprintf("%04d", i))
		pairs = append(pairs, k)
		pairs = append(pairs, k)
	}
	t5 := time.Now()
	Sets(file, pairs)
	t6 := time.Now()
	fmt.Printf("The 100 Sets took %v to run.\n", t6.Sub(t5))

	t7 := time.Now()
	Keys(file, nil, 0, 0, false)
	t8 := time.Now()
	fmt.Printf("The 100 Keys took %v to run.\n", t8.Sub(t7))

	t9 := time.Now()
	keys, _ := Keys(file, nil, 0, 0, false)
	t10 := time.Now()
	fmt.Printf("The second 100 Keys took %v to run.\n", t10.Sub(t9))

	t11 := time.Now()
	result := Gets(file, keys)
	_ = result
	//for _, r := range result {
	//log.Println(string(r))
	//}
	t12 := time.Now()
	fmt.Printf("The 100 Gets took %v to run.\n", t12.Sub(t11))
	CloseAll()
}

func TestSet(t *testing.T) {
	var err error
	//_, err = Open("1.db")
	//ch(err, t)
	defer CloseAll()
	val, err := Get("test/nodb.db", []byte("1"))
	if val != nil {
		t.Error("not nil")
	}
	//logg(val)

	err = Set("test/1.db", []byte("1"), []byte("11"))
	ch(err, t)
	err = Set("test/1.db", []byte("2"), []byte("22"))
	ch(err, t)
}

func TestGet(t *testing.T) {
	defer CloseAll()
	Set("test/1.db", []byte("1"), []byte("11"))
	res, err := Get("test/1.db", []byte("1"))
	if err != nil {
		t.Error(err)
	}
	_ = res
	//logg("Get:" + string(res))
	res2, err2 := Get("test/1.db", []byte("2"))
	if err2 != nil {
		t.Error(err2)
	}
	_ = res2
	//logg("Get:" + string(res2))
	keys, err := Keys("test/1.db", nil, 0, 0, true)
	ch(err, t)

	result := Gets("test/1.db", keys)
	_ = result
	//logg(result)
}

func TestDelete(t *testing.T) {
	var err error
	f := "test/2.db"
	DeleteFile(f)
	_, err = Open(f)
	ch(err, t)
	defer Close(f)
	err = Set(f, []byte("1"), []byte("11"))
	ch(err, t)
	err = Set(f, []byte("2"), []byte("22"))
	ch(err, t)
	res, err := Get(f, []byte("2"))

	logg(res)
	deleted, err := Delete(f, []byte("2"))
	logg(deleted)
	if !deleted {
		t.Error("not deleted")
	}
	_, err = Get(f, []byte("2"))
	logg(err)
	Close(f)
	_, err = Open(f)
	ch(err, t)
	_, err = Get(f, []byte("2"))
	logg(err)
	d, _ := Get(f, []byte("1"))
	logg(d)
	Close(f)
}

func TestRewriteVal(t *testing.T) {
	var err error
	f := "test/TestRewriteVal.db"
	//fmt.Println("123")
	DeleteFile(f)
	_, err = Open(f)
	ch(err, t)
	defer CloseAll()

	ch(Set(f, []byte("key1"), []byte("val1")), t)
	ch(Set(f, []byte("key1"), []byte("val2")), t)
	ch(Set(f, []byte("key3"), []byte("val3")), t)
	ch(Set(f, []byte("key1"), []byte("val0")), t)
	ch(Set(f, []byte("key1"), []byte("val")), t)
	v, _ := Get(f, []byte("key1"))
	//logg(string(v))
	if !bytes.Equal([]byte("val"), v) {
		t.Error("not equal")
	}
}

func TestKeys(t *testing.T) {
	var err error
	f := "test/keys.db"
	DeleteFile(f)
	_, err = Open(f)
	ch(err, t)
	defer Close(f)
	append := func(i int) {

		k := []byte(fmt.Sprintf("%02d", i))
		v := []byte("Val:" + strconv.Itoa(i))
		err := Set(f, k, v)
		ch(err, t)

	}
	for i := 1; i <= 20; i++ {
		append(i)
	}

	//ascending
	res, err := Keys(f, nil, 0, 0, true)
	var s = ""
	for _, r := range res {
		s += string(r)
	}
	if s != "0102030405060708091011121314151617181920" {
		t.Error("not asc", s)
	}
	//descending
	resdesc, err := Keys(f, nil, 0, 0, false)
	s = ""
	for _, r := range resdesc {
		s += string(r)
	}
	if s != "2019181716151413121110090807060504030201" {
		t.Error("not desc")
	}

	//offset limit asc
	reslimit, err := Keys(f, nil, 2, 2, true)
	s = ""
	for _, r := range reslimit {
		s += string(r)
	}
	if s != "0304" {
		t.Error("not off", s)
	}

	//offset limit desc
	reslimitdesc, err := Keys(f, nil, 2, 2, false)
	s = ""
	for _, r := range reslimitdesc {
		s += string(r)
	}
	if s != "1817" {
		t.Error("not off desc", s)
	}

	//from byte asc
	resfromasc, err := Keys(f, []byte("10"), 2, 2, true)
	s = ""
	for _, r := range resfromasc {
		s += string(r)
	}
	if s != "1314" {
		t.Error("not off desc", s)
	}

	//from byte desc
	resfromdesc, err := Keys(f, []byte("10"), 2, 2, false)
	s = ""
	for _, r := range resfromdesc {
		s += string(r)
	}
	if s != "0706" {
		t.Error("not off desc", s)
	}

	//from byte desc
	resnotfound, err := Keys(f, []byte("100"), 2, 2, false)
	s = ""
	for _, r := range resnotfound {
		s += string(r)
	}
	if s != "" {
		t.Error("resnotfound", s)
	}

	//from byte not eq
	resnoteq, err := Keys(f, []byte("33"), 2, 2, false)
	s = ""
	for _, r := range resnoteq {
		s += string(r)
	}
	if s != "" {
		t.Error("resnoteq", s)
	}

	//by prefix
	respref, err := Keys(f, []byte("2*"), 2, 0, false)
	s = ""
	for _, r := range respref {
		s += string(r)
	}
	if s != "20" {
		t.Error("respref", s)
	}

	//by prefix2
	respref2, err := Keys(f, []byte("1*"), 2, 0, false)
	s = ""
	for _, r := range respref2 {
		s += string(r)
	}
	if s != "1918" {
		t.Error("respref2", s)
	}

	//by prefixasc
	resprefasc, err := Keys(f, []byte("1*"), 2, 0, true)
	s = ""
	for _, r := range resprefasc {
		s += string(r)
	}
	if s != "1011" {
		t.Error("resprefasc", s, err)
	}

	//by prefixasc2
	resprefasc2, err := Keys(f, []byte("1*"), 0, 0, true)
	s = ""
	for _, r := range resprefasc2 {
		s += string(r)
	}
	if s != "10111213141516171819" {
		t.Error("resprefasc2", s, err)
	}
}

func TestWriteRead(t *testing.T) {
	len := 500
	file := "test/async.db"
	var wg sync.WaitGroup
	//var mutex = &sync.RWMutex{}
	//DeleteFile(file)
	defer CloseAll()
	append := func(i int) {
		defer wg.Done()
		//mutex.Lock()
		k := ("Key:" + strconv.Itoa(i))
		v := ("Val:" + strconv.Itoa(i))
		err := Set(file, []byte(k), []byte(v))
		//mutex.Unlock()
		if err != nil {
			t.Error(err)
		}
		//fmt.Println("Set:" + strconv.Itoa(i))
	}
	_ = append

	read := func(i int) {
		defer wg.Done()
		k := ("Key:" + strconv.Itoa(i))
		//mutex.Lock()
		b, e := Get(file, []byte(k))
		need := []byte("Val:" + strconv.Itoa(i))
		if e == nil && !bytes.Equal(need, b) {
			t.Error("Not Eq")
		}
		//fmt.Println("Get:" + strconv.Itoa(i) + " =" + string(b))
		//mutex.Unlock()
	}
	_ = read
	Open(file)
	for i := 1; i <= len; i++ {
		wg.Add(2)
		go append(i)
		go read(i)
	}

	wg.Wait()

}

func TestBinGob(t *testing.T) {
	file := "test/gob.db"
	DeleteFile(file)
	defer CloseAll()

	type Post struct {
		Id       uint32
		Content  string
		Category string
	}
	for i := 1; i < 2; i++ {
		post := &Post{Id: uint32(i), Content: "Content:" + strconv.Itoa(i)}
		err := SetGob(file, post.Id, post)
		ch(err, t)

	}

	SetGob(file, uint32(3), "post3")
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(3))
	//fmt.Println(b)
	SetGob(file, b, "post4")
	var val string
	GetGob(file, b, &val)
	//log.Println("gob!", e, val)
	if val != "post4" {
		t.Error("not post4", val)
	}

	bb := make([]byte, 4)
	binary.BigEndian.PutUint32(bb, uint32(3))
	GetGob(file, b, &val)
	if val != "post4" {
		t.Error("not post4", val)
	}
}

func TestGob(t *testing.T) {
	file := "test/gob.db"
	DeleteFile(file)
	defer CloseAll()
	type Post struct {
		Id       int
		Content  string
		Category string
	}

	for i := -4; i < 20; i++ {
		post := &Post{Id: i, Content: "Content:" + strconv.Itoa(i)}
		err := SetGob(file, post.Id, post)
		ch(err, t)
	}

	for i := -4; i < 20; i++ {
		var post = new(Post)
		err := GetGob(file, (i), post)
		ch(err, t)
		//fmt.Println("i:", i, "Post:", post)
	}
	// mix gob with other methods
	keys, err := Keys(file, nil, 1, 0, false)
	ch(err, t)
	//fmt.Println(keys)
	var k int
	buf := bytes.Buffer{}
	buf.Write(keys[0])
	if err := gob.NewDecoder(&buf).Decode(&k); err == nil {
		if k != 19 {
			t.Error("not 19")
		}
	} else {
		t.Error(err)
	}

	bin, err := Get(file, keys[0])
	ch(err, t)
	buf.Write(bin)
	p := &Post{}
	if err := gob.NewDecoder(&buf).Decode(&p); err == nil {
		//fmt.Println(p)
		if p.Id != 19 {
			t.Error("gob not 19")
		}
	} else {
		t.Error(err)
	}

	keysAsc, _ := Keys(file, nil, 0, 0, true)

	for _, v := range keysAsc {
		buf.Write(v)
		var ks int
		if err := gob.NewDecoder(&buf).Decode(&ks); err == nil {
			//с сортировкой отрицательных чисел будет хрень кнчн
			//fmt.Println(ks)
			//0,-1,1,-2,2...
		}
	}

}

func Prepend(items []interface{}, item interface{}) []interface{} {
	return append([]interface{}{item}, items...)
}
func TestSortedInsert(t *testing.T) {
	size := 10000
	var keys = make([][]byte, 0)
	var keysSort = make([][]byte, 0)
	ins := func(b []byte) {
		keysLen := len(keys)
		found := sort.Search(keysLen, func(i int) bool {
			return bytes.Compare(keys[i], b) >= 0
		})
		if found == 0 {
			//prepend
			keys = append([][]byte{b}, keys...)

		} else {
			if found >= keysLen {
				//not found - postpend ;)
				keys = append(keys, b)
			} else {
				//found
				//https://blog.golang.org/go-slices-usage-and-internals
				keys = append(keys, nil)           //grow origin slice capacity if needed
				copy(keys[found+1:], keys[found:]) //ha-ha, lol, 20x faster
				keys[found] = b
			}
		}
	}
	//ins(nil)
	for i := size; i >= 0; i-- {
		s1 := rand.NewSource(time.Now().UnixNano())
		r := rand.New(s1)
		i := r.Intn(42)

		k := []byte(fmt.Sprintf("%04d", i))

		//ins(k)
		keysSort = append(keysSort, k)
	}

	t5 := time.Now()
	_ = ins
	for _, v := range keysSort {
		//_ = v
		ins(v)
	}
	t6 := time.Now()
	fmt.Printf("The %d Sorted insert took %v to run.\n", size, t6.Sub(t5))

	//10000 insert- 1s :(
	t1 := time.Now()
	sort.Slice(keysSort, func(i, j int) bool {
		return bytes.Compare(keysSort[i], keysSort[j]) <= 0
	})
	t2 := time.Now()
	fmt.Printf("The %d Sort took %v to run.\n", size, t2.Sub(t1))
	//10000 sort - 1.360s // 8.265034ms
	//insert faster :)

	t3 := time.Now()
	sort.Slice(keysSort, func(i, j int) bool {
		return bytes.Compare(keysSort[i], keysSort[j]) <= 0
	})
	t4 := time.Now()
	fmt.Printf("The %d 2 Sort took %v to run.\n", size, t4.Sub(t3))

	for k, v := range keysSort {
		//fmt.Println(k, string(v), string(keys[k]))
		if string(v) != string(keys[k]) {
			t.Error("keys != keyssorted")
		}
	}
}

func TestPut(t *testing.T) {
	f := "test/TestPut.db"
	DeleteFile(f)
	key := []byte("1")
	err := Put(f, key, key)
	ch(err, t)
	b, err := Get(f, key)
	if !bytes.Equal(b, key) {
		t.Error("not equal")
	}
	Close(f)
	DeleteFile(f)
}

func TestHasCount(t *testing.T) {
	f := "test/TestHas.db"
	DeleteFile(f)
	key := []byte("1")
	notexist, err := Has(f, key)
	if notexist == true {
		t.Error("Has return exist", err)
	}

	zero, _ := Count(f)
	if zero != 0 {
		t.Error("Not zero")
	}
	Put(f, key, key)
	one, _ := Count(f)
	if one != 1 {
		t.Error("Not one")
	}
	exist, err := Has(f, key)
	if exist == false {
		t.Error("Has return not exist", err)
	}
	Delete(f, key)
	Close(f)
	noone, _ := Count(f)
	if noone != 0 {
		t.Error("Not one", noone)
	}
	DeleteFile(f)
}

func TestCounter(t *testing.T) {
	f := "test/TestCnt.db"
	var counter uint64
	var err error
	DeleteFile(f)
	key := []byte("postcounter")
	for i := 0; i < 10; i++ {
		counter, err = Counter(f, key)
		ch(err, t)
	}
	Close(f)
	for i := 0; i < 10; i++ {
		counter, err = Counter(f, key)
		ch(err, t)
	}
	if counter != 20 {
		t.Error("counter!=20")
	}
	key2 := []byte("counter2")
	for i := 0; i < 5; i++ {
		counter, err = Counter(f, key2)
		ch(err, t)
	}
	Close(f)
	for i := 0; i < 5; i++ {
		counter, err = Counter(f, key2)
		ch(err, t)
	}
	if counter != 10 {
		t.Error("counter!=10")
	}
}

func TestAsterix(t *testing.T) {
	f := "test/TestAsterix.db"
	DeleteFile(f)
	k := [...]string{"ka1", "ka2", "ka3", "kb1", "kb2", "k*1", "k*2", "k*3", "kn1", "kn2"}
	for _, v := range k {
		Set(f, []byte(v), nil)
	}
	k1, _ := Keys(f, []byte("ka"), uint32(0), uint32(0), true)
	_ = k1
	//fmt.Println(k1) //[]
	ka, _ := Keys(f, []byte("ka*"), uint32(0), uint32(0), true)
	res1 := ""
	for _, s := range ka {
		res1 += (string(s))
	}
	if res1 != "ka1ka2ka3" {
		t.Error("Not", "ka1ka2ka3")
	}

	kast, _ := Keys(f, []byte("k**"), uint32(0), uint32(0), true)

	res2 := ""
	for _, s := range kast {
		res2 += (string(s))
	}
	if res2 != "k*1k*2k*3" {
		t.Error("Not", "k*1k*2k*3")
	}

	//fmt.Println("kfrom")
	kfrom, _ := Keys(f, []byte("k*1"), uint32(0), uint32(0), true)
	for _, s := range kfrom {
		_ = s
		//fmt.Println(string(s))
	}
	/*
		k*2
		k*3
		ka1
		ka2
		ka3
		kb1
		kb2
		kn1
		kn2
	*/
	Close(f)
}
