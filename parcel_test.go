package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	// подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавляет новую посылку в БД
	id, err := store.Add(parcel)
	//убеждаемся в отсутствии ошибки и проверяем наличие индефикатора
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	parcel.Number = id

	// get
	// получаем только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверяем, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	stored, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, parcel, stored)

	// delete
	// удаляем добавленную посылку, убедитесь в отсутствии ошибки
	// проверяем, что посылку больше нельзя получить из БД
	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err)
	assert.Equal(t, err, sql.ErrNoRows)

}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	// подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавим новую посылку в БД, убедимся в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	// set address
	// обновим адрес, убедимся в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)
	// check
	// получим добавленную посылку и убедитесь, что адрес обновился
	checkUpdate, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, newAddress, checkUpdate.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	// подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавим новую посылку в БД, убедимся в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	// set status
	// обновим статус, убедимся в отсутствии ошибки
	err = store.SetStatus(id, ParcelStatusDelivered)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	checkUpdate, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, ParcelStatusDelivered, checkUpdate.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	// подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i]) // добавим новую посылку в БД, убедимся в отсутствии ошибки и наличии идентификатора
		require.NoError(t, err)
		assert.NotEmpty(t, id)
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	// убедимся в отсутствии ошибки
	require.NoError(t, err)
	// убедимся, что количество полученных посылок совпадает с количеством добавленных
	require.Equal(t, len(storedParcels), len(parcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		assert.Equal(t, parcel, parcelMap[parcel.Number])
	}
}
