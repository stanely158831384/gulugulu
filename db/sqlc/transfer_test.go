package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/techschool/simplebank/util"
)

func CreateAccountsAndTransfers(t *testing.T) []Account{
	user := createRandomUser(t)
	arg1 := CreateAccountParams{
		Owner: user.Username,
		Balance: util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
	account1, err1 := testStore.CreateAccount(context.Background(),arg1)
	require.NoError(t, err1)

	user2 := createRandomUser(t)
	arg2 := CreateAccountParams{
		Owner: user2.Username,
		Balance: util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
	account2, err1 := testStore.CreateAccount(context.Background(),arg2)
	require.NoError(t, err1)
	
	var accountS = []Account{account1, account2}



	return accountS 
}

func CreateRandomTransfer(t *testing.T) Transfer{

	accountArray :=  CreateAccountsAndTransfers(t)

	arg := CreateTransferParams{
		FromAccountID: accountArray[0].ID,
		ToAccountID: accountArray[1].ID,
		Amount: util.RandomAmount(),
	}

	transfer, err := testStore.CreateTransfer(context.Background(),arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)

	require.NotZero(t,transfer.ID)
	require.NotZero(t,transfer.CreatedAt)

	return transfer
}

func TestCreateTransfer(t *testing.T){
	CreateRandomTransfer(t)
}

func TestGetTransfer(t *testing.T) {
	transfer1 := CreateRandomTransfer(t)
	transfer2, err := testStore.GetTransfer(context.Background(),transfer1.ID)
	require.NoError(t,err)
	require.NotEmpty(t,transfer2)

	require.Equal(t,transfer1.ID,transfer2.ID)
	require.Equal(t,transfer1.FromAccountID,transfer2.FromAccountID)
	require.Equal(t,transfer1.ToAccountID,transfer2.ToAccountID)
	require.Equal(t,transfer1.Amount,transfer2.Amount)
	require.WithinDuration(t,transfer1.CreatedAt,transfer2.CreatedAt,time.Second)
}

func TestUpdateTransfer(t *testing.T) {
	transfer1 := CreateRandomTransfer(t)

	arg := UpdateTransferParams{
		ID: transfer1.ID,
		Amount: util.RandomAmount(),
	}

	transfer2, err := testStore.UpdateTransfer(context.Background(),arg)
	require.NoError(t,err)
	require.NotEmpty(t,transfer2)

	require.Equal(t,transfer1.ID,transfer2.ID)
	require.Equal(t,arg.Amount,transfer2.Amount)
	require.WithinDuration(t,transfer1.CreatedAt,transfer2.CreatedAt,time.Second)
}


// func TestDeleteTransfer(t *testing.T){
// 	transfer1 := CreateRandomTransfer(t)
// 	err := testStore.DeleteTransfer(context.Background(),transfer1.ID)
// 	require.NoError(t,err)

// 	transfer2, err := testStore.GetTransfer(context.Background(),transfer1.ID)
// 	require.Error(t,err)
// 	require.EqualError(t,err,db.ErrRecordNotFound.Error())
// 	require.Empty(t,transfer2)
// }

func TestListTransfers(t *testing.T){
	for i := 0; i<10; i++{
		CreateRandomTransfer(t)
	}

	arg := ListTransfersParams{
		Limit: 5,
		Offset: 5, 
	}

	transfers, err := testStore.ListTransfers(context.Background(), arg)
	require.NoError(t,err)
	require.Len(t,transfers,5)

	for _, account := range transfers {
		require.NotEmpty(t,account)
	}
}