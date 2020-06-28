package extract

import (
	"bufio"
	"ddjj/parser/declaration"
	"log"
	"strconv"
	"strings"
)

// Vehicles returns the public official's vehicles.
func Vehicles(scanner *bufio.Scanner) []*declaration.Vehicle {
	var skip = []string{
		"#",
		"TIPO VEHÍCULO",
		"MARCA",
		"MODELO",
		"CHASIS",
		"CHAPA:",
		"IMPORTE",
		"DATOS PROTEGIDOS",
	}

	scanner = moveUntil(scanner, "1.5 VEHÍCULOS", true)

	var vehicles []*declaration.Vehicle
	opts := &vehicleOpts{
		vehicle: &declaration.Vehicle{},
		counter: 0,
	}

	index := 1
	skip = append(skip, strconv.Itoa(index))
	var total int64
	for scanner.Scan() {
		line := scanner.Text()

		// Stop looking for deposits in the page when this is found.
		if line == "TOTAL VEHÍCULOS:" {
			total = getTotalInCategory(scanner)

			// Next page or end.
			scanner = moveUntil(scanner, "TIPO VEHÍCULO", true)
			line = scanner.Text()
			if line == "" {
				break
			}

			opts.importes = make([]int64, 0)
			opts.importesExtracted = false
			opts.importesIndex = 0
			opts.afterImportes = 0
			opts.counter = 0
			index = 1
		}

		if strings.Contains(line, "OBS:") {
			continue
		}
		if contains(skip, line) || line == "" {
			continue
		}

		var v *declaration.Vehicle

		// El primer item tien algunos datos extraidos antes del importe
		if index == 1 {
			v = getVehicle1(opts, line)
		} else if !opts.importesExtracted {
			// Despues del primer item, vienen los importes de todos los demás items.
			if isNumber(line) {
				opts.importes = append(opts.importes, stringToInt64(line))

				continue
			} else {
				opts.importesExtracted = true
			}
		}

		if opts.importesExtracted {
			switch opts.afterImportes {
			case 0:
				year := strings.TrimPrefix(line, "AÑO ADQUIS.: ")
				opts.firstIndex.Adquisicion = stringToYear(year)
				opts.afterImportes++
				break
			case 1:
				year := strings.TrimPrefix(line, "AÑO FABR.: ")
				opts.firstIndex.Fabricacion = stringToYear(year)
				opts.afterImportes++
				opts.counter = -1
				break
			default:
				v = getVehicle2(opts, line)
			}
		}

		if v != nil {
			vehicles = append(vehicles, v)
			opts.counter = -1
			opts.vehicle = &declaration.Vehicle{}

			if index == 1 {
				opts.firstIndex = v
			}

			if index != 1 || !strings.Contains(v.Marca, "AÑO ADQUIS.:") {
				// Skip the following item #.
				index++
				skip = append(skip, strconv.Itoa(index))
			}
		}

		opts.counter++
	}

	totalDeposits := addVehicles(vehicles)
	if totalDeposits != total {
		log.Fatal("Vehicles do not match")
	}

	return vehicles
}

type vehicleOpts struct {
	vehicle    *declaration.Vehicle
	firstIndex *declaration.Vehicle

	importes          []int64
	importesIndex     int
	importesExtracted bool
	afterImportes     int

	counter int
}

func getVehicle1(opts *vehicleOpts, line string) *declaration.Vehicle {
	switch opts.counter {
	case 0:
		opts.vehicle.Tipo = line
		break
	case 1:
		opts.vehicle.Marca = line
		break
	case 2:
		opts.vehicle.Modelo = line
		break
	case 3:
		if isNumber(line) {
			opts.vehicle.Importe = stringToInt64(line)

			return opts.vehicle
		}

		// En algunos casos como los de Justo Zacarias, 2014, todos los datos del
		// primer item se extraían antes de todos los importes. En esos casos,
		// lo que se ingreso como la marca, en realidad es el año de adquisicion
		year := strings.TrimPrefix(opts.vehicle.Marca, "AÑO ADQUIS.: ")

		// En otros casos, como en el de Gonzales Daher, 2016, el orden era correcto.
		// Entonces si la operacion de arriba cambió el valor, el orden no era correcto.
		if year != opts.vehicle.Marca {
			opts.vehicle.Adquisicion = stringToYear(year)

			opts.vehicle.Marca = opts.vehicle.Modelo
			opts.vehicle.Modelo = line
			break
		}

		year = strings.TrimPrefix(line, "AÑO ADQUIS.: ")
		opts.vehicle.Adquisicion = stringToYear(year)
		break
	case 4:
		year := strings.TrimPrefix(line, "AÑO FABR.: ")
		opts.vehicle.Fabricacion = stringToYear(year)
	case 5:
		opts.vehicle.Importe = stringToInt64(line)
		opts.afterImportes = 2

		return opts.vehicle
	}

	return nil
}

func getVehicle2(opts *vehicleOpts, line string) *declaration.Vehicle {
	switch opts.counter {
	case 0:
		opts.vehicle.Tipo = line
		break
	case 1:
		year := strings.TrimPrefix(line, "AÑO ADQUIS.: ")
		opts.vehicle.Adquisicion = stringToYear(year)
	case 2:
		opts.vehicle.Marca = line
		break
	case 3:
		// En algunos casos, el año de fabricacion es extraido primero
		// Vea Juan Afara, 2014, Mercedes Benz Camion Tumba.
		if strings.Contains(line, "AÑO FABR.: ") {
			year := strings.TrimPrefix(line, "AÑO FABR.: ")
			opts.vehicle.Fabricacion = stringToYear(year)

			break
		}

		opts.vehicle.Modelo = line
		break
	case 4:
		if opts.vehicle.Modelo == "" {
			opts.vehicle.Modelo = line
		} else {
			year := strings.TrimPrefix(line, "AÑO FABR.: ")
			opts.vehicle.Fabricacion = stringToYear(year)
		}

		opts.vehicle.Importe = opts.importes[opts.importesIndex]
		opts.importesIndex++

		return opts.vehicle
	}

	return nil
}

func addVehicles(vehicles []*declaration.Vehicle) int64 {
	var total int64
	for _, v := range vehicles {
		total += v.Importe
	}

	return total
}
