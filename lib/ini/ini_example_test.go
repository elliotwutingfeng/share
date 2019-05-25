package ini

import (
	"fmt"
	"log"
	"os"
)

func ExampleIni_Add() {
	ini := new(Ini)

	ini.Add("", "", "k1", "v1")
	ini.Add("s1", "", "", "v2")

	ini.Add("s1", "", "k1", "")
	ini.Add("s1", "", "k1", "v1")
	ini.Add("s1", "", "k1", "v2")
	ini.Add("s1", "", "k1", "v1")

	ini.Add("s1", "sub", "k1", "v1")
	ini.Add("s1", "sub", "k1", "v1")

	ini.Add("s2", "sub", "k1", "v1")

	err := ini.Write(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// [s1]
	// k1 = true
	// k1 = v1
	// k1 = v2
	// [s1 "sub"]
	// k1 = v1
	// [s2 "sub"]
	// k1 = v1
}

func ExampleIni_Gets() {
	input := []byte(`
[section]
key=value1

[section "sub"]
key=value2

[section]
key=value3

[section "sub"]
key=value4
key=value2
`)

	inis, _ := Parse(input)

	fmt.Println(inis.Gets("section", "", "key"))
	fmt.Println(inis.Gets("section", "sub", "key"))
	//Output:
	//[value1 value3]
	//[value2 value4]
}

func ExampleIni_AsMap() {
	input := []byte(`
[section]
key=value1
key2=

[section "sub"]
key=value1

[section]
key=value2
key2=false

[section "sub"]
key=value2
key=value3
`)

	inis, err := Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	iniMap := inis.AsMap()

	for k, v := range iniMap {
		fmt.Println(k, "=", v)
	}
	// Unordered output:
	// section::key = [value1 value2]
	// section::key2 = [true false]
	// section:sub:key = [value1 value2 value3]
}

func ExampleIni_Prune() {
	input := []byte(`
[section]
key=value1 # comment
key2= ; another comment

[section "sub"]
key=value1

; here is comment on section
[section]
key=value2
key2=false

[section "sub"]
key=value2
key=value1
`)

	in, err := Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	in.Prune()

	for _, sec := range in.secs {
		fmt.Printf("%s", sec)
		for _, v := range sec.vars {
			fmt.Printf("%s", v)
		}
	}
	// Output:
	// [section]
	// key = value1
	// key2 = true
	// key = value2
	// key2 = false
	// [section "sub"]
	// key = value2
	// key = value1
}

func ExampleIni_Rebase() {
	input := []byte(`
		[section]
		key=value1
		key2=

		[section "sub"]
		key=value1
`)

	other := []byte(`
		[section]
		key=value2
		key2=false

		[section "sub"]
		key=value2
		key=value1
`)

	in, err := Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	in2, err := Parse(other)
	if err != nil {
		log.Fatal(err)
	}

	in.Prune()
	in2.Prune()

	in.Rebase(in2)

	for _, sec := range in.secs {
		fmt.Printf("%s", sec)
		for _, v := range sec.vars {
			fmt.Printf("%s", v)
		}
	}
	// Output:
	// [section]
	// key = value1
	// key2 = true
	// key = value2
	// key2 = false
	// [section "sub"]
	// key = value2
	// key = value1
}

func ExampleIni_Set() {
	input := []byte(`
[section]
key=value1 # comment
key2= ; another comment

[section "sub"]
key=value1

[section] ; here is comment on section
key=value2
key2=false

[section "sub"]
key=value2
key=value1
`)

	ini, err := Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	ini.Set("", "sub", "key", "value3")
	ini.Set("sectionnotexist", "sub", "key", "value3")
	ini.Set("section", "sub", "key", "value3")
	ini.Set("section", "", "key", "value4")
	ini.Set("section", "", "keynotexist", "value4")

	err = ini.Write(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	//Output:
	//[section]
	//key=value1 # comment
	//key2= ; another comment
	//
	//[section "sub"]
	//key=value1
	//
	//[section] ; here is comment on section
	//key=value4
	//key2=false
	//
	//[section "sub"]
	//key=value2
	//key=value3
}

func ExampleIni_Unset() {
	input := []byte(`
[section]
key=value1 # comment
key2= ; another comment

[section "sub"]
key=value1

; here is comment on section
[section]
key=value2
key2=false

[section "sub"]
key=value2
key=value1
`)

	ini, err := Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	ini.Unset("", "sub", "keynotexist")
	ini.Unset("sectionnotexist", "sub", "keynotexist")
	ini.Unset("section", "sub", "keynotexist")
	ini.Unset("section", "sub", "key")
	ini.Unset("section", "", "keynotexist")
	ini.Unset("section", "", "key")

	err = ini.Write(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	//Output:
	//[section]
	//key=value1 # comment
	//key2= ; another comment
	//
	//[section "sub"]
	//key=value1
	//
	//; here is comment on section
	//[section]
	//key2=false
	//
	//[section "sub"]
	//key=value2
}
