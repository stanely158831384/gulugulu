package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/techschool/simplebank/util"
)

func createRandomEntry(t *testing.T,account Account) Entry{
	arg := CreateEntryParams{
		AccountID: account.ID,
		Amount: util.RandomAmount(),
	}

	entry, err := testStore.CreateEntry(context.Background(),arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)

	require.NotZero(t,entry.ID)
	require.NotZero(t,entry.CreatedAt)

	return entry
}

func TestCreateEntry(t *testing.T){
	user := createRandomUser(t)
	arg := CreateAccountParams{
		Owner: user.Username,
		Balance: util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
	account, err := testStore.CreateAccount(context.Background(),arg)
	require.NoError(t,err)
	createRandomEntry(t,account)
}

func TestGetEntry(t *testing.T) {
	user := createRandomUser(t)
	arg := CreateAccountParams{
		Owner: user.Username,
		Balance: util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
	account, err := testStore.CreateAccount(context.Background(),arg)

	entry1 := createRandomEntry(t,account)
	entry2, err := testStore.GetEntry(context.Background(),entry1.ID)
	require.NoError(t,err)
	require.NotEmpty(t,entry2)

	require.Equal(t,entry1.ID,entry2.ID)
	require.Equal(t,entry1.AccountID,entry2.AccountID)
	require.Equal(t,entry1.Amount,entry2.Amount)
	require.WithinDuration(t,entry1.CreatedAt,entry2.CreatedAt,time.Second)
}

func TestUpdateEntry(t *testing.T) {
	user := createRandomUser(t)
	arg1 := CreateAccountParams{
		Owner: user.Username,
		Balance: util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
	account, err := testStore.CreateAccount(context.Background(),arg1)
	require.NoError(t,err)

	account1 := createRandomEntry(t,account)

	arg := UpdateEntryParams{
		ID: account1.ID,
		Amount: account1.Amount,
	}

	account2, err1 := testStore.UpdateEntry(context.Background(),arg)
	require.NoError(t,err1)
	require.NotEmpty(t,account2)

	require.Equal(t,account1.ID,account2.ID)
	require.Equal(t,account1.Amount,account2.Amount)

	require.WithinDuration(t,account1.CreatedAt,account2.CreatedAt,time.Second)
}

// func TestDeleteEntry(t *testing.T){
// 	user := createRandomUser(t)
// 	arg1 := CreateAccountParams{
// 		Owner: user.Username,
// 		Balance: util.RandomMoney(),
// 		Currency: util.RandomCurrency(),
// 	}
// 	account, err := testStore.CreateAccount(context.Background(),arg1)
// 	require.NoError(t,err)

// 	account1 := createRandomEntry(t,account)
// 	err1 := testStore.DeleteEntry(context.Background(),account1.ID)
// 	require.NoError(t,err1)

// 	account2, err := testStore.GetEntry(context.Background(),account1.ID)
// 	require.Error(t,err)
// 	require.EqualError(t,err,db.ErrRecordNotFound.Error())
// 	require.Empty(t,account2)
// }
 

func TestListEntries(t *testing.T){
	account := createRandomAccount(t);
	for i := 0; i<10; i++{
		createRandomEntry(t,account)
	}

	arg := ListEntriesParams{
		Limit: 5,
		Offset: 5, 
	}

	entries, err := testStore.ListEntries(context.Background(), arg)
	require.NoError(t,err)
	require.Len(t,entries,5)

	for _, entry := range entries {
		require.NotEmpty(t,entry)
	}
}

// func TestListEntries(t *testing.T) {
// 	account := createRandomAccount(t)
// 	for i := 0; i < 10; i++ {
// 		createRandomEntry(t, account)
// 	}

// 	arg := ListEntriesParams{
// 		AccountID: account.ID,
// 		Limit:     5,
// 		Offset:    5,
// 	}

// 	entries, err := testStore.ListEntries(context.Background(), arg)
// 	require.NoError(t, err)
// 	require.Len(t, entries, 5)

// 	for _, entry := range entries {
// 		require.NotEmpty(t, entry)
// 		require.Equal(t, arg.AccountID, entry.AccountID)
// 	}
// }