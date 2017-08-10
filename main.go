package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"sort"
)

type VariantInfo struct {
	String *string `toml:"string"`
}

type VariantName string
type EnumVariantMap map[string]VariantInfo

type EnumName string
type EnumInfo struct {
	PrefixName string         `toml:"prefix"`
	Variants   EnumVariantMap `toml:"variants"`
}
type EnumTypeMap map[string]EnumInfo

func nl(v string) string {
	v += "\n"
	return v
}

func generateCode(m EnumTypeMap, enable_json bool, enable_bson bool) string {
	var code string
	var sorted_list = sort.StringSlice{}
	for enum_name := range m {
		sorted_list = append(sorted_list, enum_name)
	}
	sorted_list.Sort()

	for _, enum_name := range sorted_list {
		var enum_info = m[enum_name]

		var sorted_var_list = sort.StringSlice{}
		for v := range enum_info.Variants {
			sorted_var_list = append(sorted_var_list, v)
		}
		sorted_var_list.Sort()

		var full_names = map[string]string{}
		for var_name := range enum_info.Variants {
			full_names[var_name] = fmt.Sprintf("%s%s", enum_info.PrefixName, var_name)
		}

		code += nl(fmt.Sprintf("type %sIface interface { \n %sIfaceFunc() }", enum_name, enum_name))
		code += nl(fmt.Sprintf("type %s struct { %sIface }", enum_name, enum_name))
		code += nl(fmt.Sprintf("func (self *%s) Value() %sIface { return self.%sIface; }", enum_name, enum_name, enum_name))

		for _, var_name := range sorted_var_list {
			var full_var_name = full_names[var_name]
			code += nl(fmt.Sprintf("type %s struct{}", full_var_name))
			code += nl(fmt.Sprintf("func (%s) %sIfaceFunc() {}", full_var_name, enum_name))
		}

		var stringer_code = func() *string {
			var case_code = func() *string {
				var case_code string
				var sorted_list = sort.StringSlice{}
				for v := range enum_info.Variants {
					sorted_list = append(sorted_list, v)
				}
				sorted_list.Sort()
				for _, var_name := range sorted_list {
					case_code += nl(fmt.Sprintf("case %s:", full_names[var_name]))
					case_code += nl(fmt.Sprintf("	return \"%s\"", func() string {
						var v = enum_info.Variants[var_name]
						if v.String != nil {
							return *v.String
						} else {
							return full_names[var_name]
						}
					}()))
				}

				return &case_code
			}()

			if case_code == nil {
				return nil
			}
			var v = nl(fmt.Sprintf("switch self.%sIface.(type) { \n %s \n } \n panic(errors.New(\"Invalid value of %s\"))", enum_name, *case_code, enum_name))
			return &v
		}()
		if stringer_code != nil {
			code += nl(fmt.Sprintf("func (self *%s) String() string { \n %s \n }", enum_name, *stringer_code))
		}

		if enable_json {
			var marshal_code = func() *string {
				var case_code = func() *string {
					var case_code string
					var sorted_list = sort.StringSlice{}
					for v := range enum_info.Variants {
						sorted_list = append(sorted_list, v)
					}
					sorted_list.Sort()
					for _, var_name := range sorted_list {
						var var_info = enum_info.Variants[var_name]
						if var_info.String == nil {
							return nil
						}

						case_code += nl(fmt.Sprintf("case %s:", full_names[var_name]))
						case_code += nl(fmt.Sprintf("	return json.Marshal(\"%s\")", *var_info.String))
					}

					return &case_code
				}()

				if case_code == nil {
					return nil
				}
				var v = nl(fmt.Sprintf("switch self.Value().(type) { \n %s \n } \n return nil, errors.New(\"Invalid value of %s\")", *case_code, enum_name))
				return &v
			}()
			if marshal_code != nil {
				code += nl(fmt.Sprintf("func (self %s) MarshalJSON() ([]byte, error) { \n %s \n }", enum_name, *marshal_code))
			}
		}

		if enable_bson {
			var marshal_code = func() *string {
				var case_code = func() *string {
					var case_code string
					var sorted_list = sort.StringSlice{}
					for v := range enum_info.Variants {
						sorted_list = append(sorted_list, v)
					}
					sorted_list.Sort()
					for _, var_name := range sorted_list {
						var var_info = enum_info.Variants[var_name]
						if var_info.String == nil {
							return nil
						}

						case_code += nl(fmt.Sprintf("case %s:", full_names[var_name]))
						case_code += nl(fmt.Sprintf("	return \"%s\", nil", *var_info.String))
					}

					return &case_code
				}()

				if case_code == nil {
					return nil
				}
				var v string
				v += nl(fmt.Sprintf("var v = self.Value(); if v == nil { return nil, errors.New(\"%s cannot be nil\") }", enum_name))
				v += nl(fmt.Sprintf("switch v.(type) { \n %s \n } \n return nil, errors.New(\"Invalid value of %s\")", *case_code, enum_name))
				return &v
			}()
			if marshal_code != nil {
				code += nl(fmt.Sprintf("func (self %s) GetBSON() (interface{}, error) { \n %s \n }", enum_name, *marshal_code))
			}
		}

		if enable_json {
			var unmarshal_code = func() *string {
				var case_code = func() *string {
					var case_code string
					var sorted_list = sort.StringSlice{}
					for v := range enum_info.Variants {
						sorted_list = append(sorted_list, v)
					}
					sorted_list.Sort()
					for _, var_name := range sorted_list {
						var var_info = enum_info.Variants[var_name]
						if var_info.String == nil {
							return nil
						}

						case_code += nl(fmt.Sprintf("case \"%s\":", *var_info.String))
						case_code += nl(fmt.Sprintf("	self.%sIface = %s{}", enum_name, full_names[var_name]))
						case_code += nl("	return nil")
					}

					return &case_code
				}()

				if case_code == nil {
					return nil
				}
				var v string
				v += nl("var s string")
				v += nl("if err := json.Unmarshal(b, &s); err != nil { return err }")
				v += nl("switch s {")
				v += nl(*case_code)
				v += nl("}")
				v += nl(fmt.Sprintf("return errors.New(\"Unknown %s\")\n", enum_name))
				return &v
			}()
			if unmarshal_code != nil {
				code += nl(fmt.Sprintf("func (self *%s) UnmarshalJSON(b []byte) error {\n %s \n}\n", enum_name, *unmarshal_code))
			}
		}

		if enable_bson {
			var unmarshal_code = func() *string {
				var case_code = func() *string {
					var case_code string
					var sorted_list = sort.StringSlice{}
					for v := range enum_info.Variants {
						sorted_list = append(sorted_list, v)
					}
					sorted_list.Sort()
					for _, var_name := range sorted_list {
						var var_info = enum_info.Variants[var_name]
						if var_info.String == nil {
							return nil
						}

						case_code += nl(fmt.Sprintf("case \"%s\":", *var_info.String))
						case_code += nl(fmt.Sprintf("	self.%sIface = %s{}", enum_name, full_names[var_name]))
						case_code += nl("	return nil")
					}

					return &case_code
				}()

				if case_code == nil {
					return nil
				}
				var v string
				v += nl("var s string")
				v += nl("if err := v.Unmarshal(&s); err != nil { return err }")
				v += nl("switch s {")
				v += nl(*case_code)
				v += nl("}")
				v += nl(fmt.Sprintf("return errors.New(\"Unknown %s\")\n", enum_name))
				return &v
			}()
			if unmarshal_code != nil {
				code += nl(fmt.Sprintf("func (self *%s) SetBSON(v bson.Raw) error { \n %s \n }", enum_name, *unmarshal_code))
			}
		}
	}

	return code
}

func main() {
	var enable_json bool = false
	flag.BoolVar(&enable_json, "enable-json", false, "Enable generation of JSON code")

	var enable_bson bool = false
	flag.BoolVar(&enable_bson, "enable-bson", false, "Enable generation of BSON code")

	flag.Parse()

	var data string
	var scanner = bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		data += scanner.Text()
		data += "\n"
	}

	var m = EnumTypeMap{}
	var err = toml.Unmarshal([]byte(data), &m)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(generateCode(m, enable_json, enable_bson))
}
