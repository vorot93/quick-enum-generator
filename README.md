# quick-enum-generator
Generate type-safe JSON-serializable Enums for your Golang project.

## Installation
`go install https://github.com/vorot93/quick-enum-generator`

## Usage
1. Create your models in ini-like [TOML](https://github.com/toml-lang/toml) format:

```
[Animal]
[Animal.variants]
Dog  = { jkey = "doggie" }
Cat  = { jkey = "horsie" }
Pony = { jkey = "littlepony" }

[CarMake]
prefix = "ManufacturedBy"
[CarMake.variants]
BMW     = { jkey = "bmw" }
Ford    = { jkey = "ford" }
Peugeot = { jkey = "peugeot" }
```

2. Run `quick-enum-generator`, input the models, hit Ctrl+D and use the resulting Golang code. Alternatively use the following command:

`quick-enum-generator -enable-json < ./enums.toml > ./enums.go` - make sure to add path to the installed executable (YOUR_GOPATH/bin) if the latter is not in PATH.

3. You have your enums. Let's use them:
```
var my_car        = CarMake{ManufacturedByFord{}}
var neighbors_car = CarMake{ManufacturedByBMW{}}

var my_car_s, _        = json.Marshal(my_car)
var neighbors_car_s, _ = json.Marshal(neighbors_car)

fmt.Printf("Is my car made by the same company? %b, mine was made by %s while neighbor's by %s", my_car == neighbors_car, my_car_s, neighbors_car_s)
```
