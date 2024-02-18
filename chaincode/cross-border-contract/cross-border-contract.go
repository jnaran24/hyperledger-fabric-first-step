package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Cliente struct {
	Id             string
	NombreCompleto string
	Saldo          float64
	ValorPromedio  float64
	Moneda         string
}

// Lista de clientes
var clientes = []Cliente{
	{
		Id:             "1020498574",
		NombreCompleto: "Diego Salazar Rojas",
		Saldo:          1000.0,
		ValorPromedio:  200.0,
		Moneda:         "EUR",
	},
	{
		Id:             "1030599585",
		NombreCompleto: "Ana García López",
		Saldo:          500.0,
		ValorPromedio:  150.0,
		Moneda:         "USD",
	},
	{
		Id:             "1040600596",
		NombreCompleto: "Juan Pérez Martínez",
		Saldo:          2000.0,
		ValorPromedio:  300.0,
		Moneda:         "COP",
	},
	{
		Id:             "1050701607",
		NombreCompleto: "María González Pérez",
		Saldo:          1500.0,
		ValorPromedio:  250.0,
		Moneda:         "EUR",
	},
	{
		Id:             "1060802618",
		NombreCompleto: "Pedro Rodríguez García",
		Saldo:          1200.0,
		ValorPromedio:  200.0,
		Moneda:         "USD",
	},
}

// Estructura de la transacción
type Transaccion struct {
	IdCliente     string    `json:"idCliente"`
	Monto         float64   `json:"monto"`
	Destino       string    `json:"destino"`
	MonedaDestino string    `json:"monedaDestino"`
	IdTransaccion string    `json:"idTransaccion"`
	Timestamp     time.Time `json:"timestamp"`
}

// Lista de entidades sancionadas
var entidadesSancionadas = []string{
	"entidadSancionada1",
	"entidadSancionada3",
	"entidadSancionada5",
}

func (s *SmartContract) CrearTransaccion(ctx contractapi.TransactionContextInterface, idCliente string, monto float64, monedaDestino string, destino string, idTransaccion string) error {
	//validaciones

	// Validar si el ID de la transacción ya existe
	transactionAsBytesQuery, err := ctx.GetStub().GetState(idTransaccion)
	if err != nil {
		return fmt.Errorf("Failed to read from world state. %s", err.Error())
	}
	if transactionAsBytesQuery != nil {
		return fmt.Errorf("%s already exist", idTransaccion)
	}

	cliente := BuscarClientePorID(idCliente)
	var currentClient Cliente = cliente

	// Validar fondos
	if monto > GetSaldo(idCliente) {
		return fmt.Errorf("Fondos insuficientes")
	}

	// Obtener la moneda del cliente
	monedaOrigen := currentClient.Moneda

	// Validar entidades sancionadas
	if EstaSancionado(destino) {
		return fmt.Errorf("Entidad sancionada")
	}

	// Validar transacción sospechosa
	if monto > currentClient.ValorPromedio*1.5 {
		return fmt.Errorf("Transacción sospechosa")
	}

	// Convertir moneda (si es necesario)
	if monedaOrigen != monedaDestino {
		monto = ConvertirMoneda(monto, monedaOrigen, monedaDestino)
	}

	// Crear transacción
	transaccion := Transaccion{
		IdCliente:     idCliente,
		Monto:         monto,
		MonedaDestino: monedaDestino,
		Destino:       destino,
		IdTransaccion: idTransaccion,
		//Firma:     firma,
		//Hash:      hash,
		Timestamp: time.Now(),
	}

	transactionAsBytes, _ := json.Marshal(transaccion)

	//save in ledger
	return ctx.GetStub().PutState(idTransaccion, transactionAsBytes)
}

func (s *SmartContract) ConsultarTransaccion(ctx contractapi.TransactionContextInterface, idTransaccion string) (*Transaccion, error) {
	transactionAsBytes, err := ctx.GetStub().GetState(idTransaccion)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if transactionAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", idTransaccion)
	}

	transaccion := new(Transaccion)

	err = json.Unmarshal(transactionAsBytes, transaccion)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal error. %s", err.Error())
	}

	return transaccion, nil
}

func BuscarClientePorID(id string) Cliente {
	for _, cliente := range clientes {
		if cliente.Id == id {
			return cliente
		}
	}

	return Cliente{}
}

// Función para convertir moneda
func ConvertirMoneda(monto float64, monedaOrigen string, monedaDestino string) float64 {
	var tasaCambio float64

	switch {
	case monedaOrigen == "COP" && monedaDestino == "USD":
		tasaCambio = 1.0 / 4000.0
	case monedaOrigen == "USD" && monedaDestino == "COP":
		tasaCambio = 4000.0
	case monedaOrigen == "USD" && monedaDestino == "EUR":
		tasaCambio = 0.93
	case monedaOrigen == "EUR" && monedaDestino == "USD":
		tasaCambio = 1.0 / 0.93
	case monedaOrigen == "COP" && monedaDestino == "EUR":
		tasaCambio = 1.0 / 4250.0
	case monedaOrigen == "EUR" && monedaDestino == "COP":
		tasaCambio = 4250.0
	default:
		fmt.Println("Moneda no compatible:", monedaOrigen, monedaDestino)
		return monto
	}

	return monto * tasaCambio
}

// Obtener el tipo de moneda del cliente
func GetMonedaCliente(idCliente string) string {
	for _, cliente := range clientes {
		if cliente.Id == idCliente {
			return cliente.Moneda
		}
	}
	return "" // Moneda por defecto si no se encuentra el cliente
}

// Función para verificar si una entidad está sancionada
func EstaSancionado(entidad string) bool {
	for _, sancionado := range entidadesSancionadas {
		if strings.EqualFold(entidad, sancionado) {
			return true
		}
	}
	return false
}

// Función para obtener el saldo del cliente
func GetSaldo(cliente string) float64 {
	for _, clienteEnLista := range clientes {
		if clienteEnLista.Id == cliente { // Comparar el ID del cliente con la propiedad Id del objeto en la lista
			return clienteEnLista.Saldo
		}
	}
	return 0.0 // Saldo por defecto si no se encuentra el cliente
}
func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create cross-border-contract chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error create cross-border-contract chaincode: %s", err.Error())
	}
}
