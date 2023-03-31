package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">> before:", account1.Balance, account2.Balance)

	// run concurrent transfer transactions
	n := 5
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	existed := make(map[int]bool)

	for i := 0; i < n; i++ {
		go func(i int) {
			txName  := fmt.Sprintf("tx %d", i+1)
			ctx := context.WithValue(context.Background(), txKey, txName)

			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result 
		}(i)
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <- results
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, amount, transfer.Amount)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)
		
		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)
		
		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		fmt.Println(">> tx:", fromAccount.Balance, toAccount.Balance)

		// check accounts' balance
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0) // 1 * amount, 2 * amount ...

		k := int(diff1 /amount)
		require.True(t, k >=1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// check final balance
	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, updatedAccount1)

	require.Equal(t, account1.Balance - int64(n) * amount, updatedAccount1.Balance)
	
	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, updatedAccount2)

	require.Equal(t, account2.Balance + int64(n) * amount, updatedAccount2.Balance)
	
	fmt.Println(">> after:", updatedAccount1.Balance, updatedAccount2.Balance)
}

func TestTransferTxDeadLock(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">> before:", account1.Balance, account2.Balance)

	// run concurrent transfer transactions
	n := 10
	amount := int64(10)

	errs := make(chan error)


	for i := 0; i < n; i++ {
		go func(i int) {
			txName  := fmt.Sprintf("tx %d", i+1)
			ctx := context.WithValue(context.Background(), txKey, txName)

			fromAccountID := account1.ID
			toAccountID := account2.ID

			if i % 2 == 1 {
				fromAccountID = account2.ID
				toAccountID = account1.ID
			}

			_, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			errs <- err
		}(i)
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	// check final balance
	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, updatedAccount1)

	require.Equal(t, account1.Balance, updatedAccount1.Balance)
	
	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, updatedAccount2)

	require.Equal(t, account2.Balance, updatedAccount2.Balance)
	
	fmt.Println(">> after:", updatedAccount1.Balance, updatedAccount2.Balance)
}
