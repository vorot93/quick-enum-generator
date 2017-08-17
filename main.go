package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"sort"
)

type EnumVariantMap map[string]string

type EnumInfo struct {
	Default  *string        `toml:"default"`
	Variants EnumVariantMap `toml:"variants"`
}
type EnumTypeMap map[string]EnumInfo

func nl(v string) string {
	v += "\n"
	return v
}

func set_default(enum_name string, enum_info EnumInfo) string {
	if enum_info.Default != nil {
		return fmt.Sprintf("if len(*self) == 0 { *self = %s }", enum_name+*enum_info.Default)
	} else {
		return ""
	}
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
			full_names[var_name] = fmt.Sprintf("%s%s", enum_name, var_name)
		}

		code += nl(fmt.Sprintf("type %s string", enum_name))

		for _, var_name := range sorted_var_list {
			var full_var_name = full_names[var_name]
			code += nl(fmt.Sprintf("const %s %s = \"%s\"", full_var_name, enum_name, enum_info.Variants[var_name]))
		}

		code += nl(fmt.Sprintf("func (self %s) GetPtr() *%s { var v = self; return &v; }", enum_name, enum_name))

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
						if len(v) > 0 {
							return v
						} else {
							return var_name
						}
					}()))
				}

				return &case_code
			}()

			if case_code == nil {
				return nil
			}
			var v = nl(fmt.Sprintf("switch *self { \n %s \n } \n panic(errors.New(\"Invalid value of %s\"))", *case_code, enum_name))
			return &v
		}()
		if stringer_code != nil {
    			code += "\n"
			code += nl(fmt.Sprintf("func (self *%s) String() string { %s %s \n }", enum_name, nl(set_default(enum_name, enum_info)), nl(*stringer_code)))
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
						var s = enum_info.Variants[var_name]

						case_code += nl(fmt.Sprintf("case %s:", full_names[var_name]))
						case_code += nl(fmt.Sprintf("	return json.Marshal(\"%s\")", s))
					}

					return &case_code
				}()

				if case_code == nil {
					return nil
				}
				var v = nl(fmt.Sprintf("switch *self { \n %s \n } \n return nil, errors.New(\"Invalid value of %s\")", *case_code, enum_name))
				return &v
			}()
			if marshal_code != nil {
    				code += "\n"
				code += nl(fmt.Sprintf("func (self *%s) MarshalJSON() ([]byte, error) { %s %s \n }", enum_name, nl(set_default(enum_name, enum_info)), nl(*marshal_code)))
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
						var s = enum_info.Variants[var_name]

						case_code += nl(fmt.Sprintf("case %s:", full_names[var_name]))
						case_code += nl(fmt.Sprintf("	return \"%s\", nil", s))
					}

					return &case_code
				}()

				if case_code == nil {
					return nil
				}
				var v string
				v += nl(fmt.Sprintf("switch *self { \n %s \n } \n return nil, errors.New(\"Invalid value of %s\")", *case_code, enum_name))
				return &v
			}()
			if marshal_code != nil {
    				code += "\n"
				code += nl(fmt.Sprintf("func (self *%s) GetBSON() (interface{}, error) { %s %s \n }", enum_name, nl(set_default(enum_name, enum_info)), nl(*marshal_code)))
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
						var s = enum_info.Variants[var_name]

						case_code += nl(fmt.Sprintf("case \"%s\":", s))
						case_code += nl(fmt.Sprintf("	*self = %s", full_names[var_name]))
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
						var s = enum_info.Variants[var_name]

						case_code += nl(fmt.Sprintf("case \"%s\":", s))
						case_code += nl(fmt.Sprintf("	*self = %s", full_names[var_name]))
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
