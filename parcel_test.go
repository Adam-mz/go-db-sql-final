package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func TestAddGetDelete(t *testing.T) {
	// Подготовка
	db, err := sql.Open("sqlite", ":memory:") // Используем in-memory SQLite для тестов
	require.NoError(t, err, "Failed to connect to the database")
	defer db.Close()

	// Инициализация таблицы
	_, err = db.Exec(`
		CREATE TABLE parcel (
			number INTEGER PRIMARY KEY AUTOINCREMENT,
			client INTEGER NOT NULL,
			status TEXT NOT NULL,
			address TEXT NOT NULL,
			created_at TEXT NOT NULL
		)
	`)
	require.NoError(t, err, "Failed to create table")

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Добавление
	id, err := store.Add(parcel)
	require.NoError(t, err, "Failed to add parcel to the store")
	require.NotZero(t, id, "Parcel ID should not be zero")

	// Получение
	storedParcel, err := store.Get(id)
	require.NoError(t, err, "Failed to get parcel from the store")

	// Проверяем, что значения всех полей совпадают
	require.Equal(t, id, storedParcel.Number, "Parcel numbers do not match")
	require.Equal(t, parcel.Client, storedParcel.Client, "Parcel clients do not match")
	require.Equal(t, parcel.Address, storedParcel.Address, "Parcel addresses do not match")
	require.Equal(t, parcel.Status, storedParcel.Status, "Parcel statuses do not match")
	require.Equal(t, parcel.CreatedAt, storedParcel.CreatedAt, "Parcel creation times do not match")

	// Удаление
	err = store.Delete(id)
	require.NoError(t, err, "Failed to delete parcel from the store")

	// Проверяем, что запись больше не существует
	_, err = store.Get(id)
	require.ErrorIs(t, err, sql.ErrNoRows, "Expected sql.ErrNoRows when getting a deleted parcel")
}

func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	p, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, p.Address)
}

func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	p, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, p.Status)
}

func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	clientID := randRange.Intn(10_000_000)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	for i := range parcels {
		parcels[i].Client = clientID
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		parcels[i].Number = id
	}

	storedParcels, err := store.GetByClient(clientID)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	for i, sp := range storedParcels {
		require.Equal(t, parcels[i].Number, sp.Number, "Parcel numbers do not match")
		require.Equal(t, parcels[i].Client, sp.Client, "Parcel clients do not match")
		require.Equal(t, parcels[i].Address, sp.Address, "Parcel addresses do not match")
		require.Equal(t, parcels[i].Status, sp.Status, "Parcel statuses do not match")
		require.Equal(t, parcels[i].CreatedAt, sp.CreatedAt, "Parcel creation times do not match")
	}
}
