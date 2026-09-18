package domain

import (
	"context"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/leonardovalentini/crypto-coins/crypto/dto"
	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/leonardovalentini/crypto-coins/lib/logger"
)

type PriceRepositoryDb struct {
	client *sqlx.DB
}

func NewPriceRepositoryDb(client *sqlx.DB) PriceRepository {
	return &PriceRepositoryDb{client: client}
}

func (d *PriceRepositoryDb) Save(ctx context.Context, p Price) (*Price, error) {
	loggerWithCxt := logger.NewLogger(ctx)
	query := "INSERT INTO prices (coin, coin_ref, date, price, site) values (?, ?, ?, ?, ?)"

	result, err := d.client.ExecContext(ctx, query, p.Coin, p.CoinRef, p.Date, p.Price, p.Site)
	if err != nil {
		loggerWithCxt.Error("Error while creating new price: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected error from database")
	}

	id, err := result.LastInsertId()
	if err != nil {
		loggerWithCxt.Error("Error while getting last insert id for new price: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected error from database")
	}
	p.Id = int(id)
	return &p, nil
}

func (d *PriceRepositoryDb) FindAllBy(ctx context.Context, userId string, gpr *dto.GetPricesRequest) ([]Price, error) {
	var prices []Price
	offset := gpr.Page * gpr.Size

	joinClause := `
		FROM prices AS p 
		INNER JOIN user_coins AS uc 
		ON uc.user_id = ? AND p.coin = uc.coin AND p.coin_ref = uc.coin_ref `
	whereClause, whereArgs := getWhereClause(gpr)

	joinAndWhereClause := fmt.Sprintf(`%s %s`, joinClause, whereClause)
	joinAndWhereArgs := []interface{}{userId}
	joinAndWhereArgs = append(joinAndWhereArgs, whereArgs...)

	itemsQuery := fmt.Sprintf(`
		SELECT p.coin, p.coin_ref, p.date, p.price, p.site 
		%s 
		LIMIT ? OFFSET ?`, joinAndWhereClause)
	itemsArgs := append(joinAndWhereArgs, gpr.Size, offset)

	err := d.client.SelectContext(ctx, &prices, itemsQuery, itemsArgs...)
	if err != nil {
		loggerWithCxt := logger.NewLogger(ctx)
		loggerWithCxt.Error("Error while searching for prices: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected error from database")
	}

	if prices == nil {
		prices = []Price{}
	}

	return prices, err
}

func (d *PriceRepositoryDb) GetSummary(ctx context.Context, userId string, gpr *dto.GetPricesRequest) (*int, *Summary, error) {
	joinAndWhereClause, joinAndWhereArgs := getJoinAndWhereClause(userId, gpr)

	var summaryRows []SummaryRow
	summaryQuery := fmt.Sprintf(`
		SELECT "coin" AS column_name, p.coin AS value, COUNT(*) AS count
		%s 
		GROUP BY p.coin
		UNION ALL
		SELECT "coin_ref" AS columnName, p.coin_ref AS value, COUNT(*) AS count
		%s 
		GROUP BY p.coin_ref
		UNION ALL
		SELECT "site" AS columnName, p.site AS value, COUNT(*) AS count
		%s 
		GROUP BY p.site
		`, joinAndWhereClause, joinAndWhereClause, joinAndWhereClause)
	var summaryArgs []interface{}
	summaryArgs = append(summaryArgs, joinAndWhereArgs...)
	summaryArgs = append(summaryArgs, joinAndWhereArgs...)
	summaryArgs = append(summaryArgs, joinAndWhereArgs...)

	err := d.client.SelectContext(ctx, &summaryRows, summaryQuery, summaryArgs...)
	if err != nil {
		loggerWithCxt := logger.NewLogger(ctx)
		loggerWithCxt.Error("Error while searching for prices: " + err.Error())
		return nil, nil, errs.NewUnexpectedError("Unexpected error from database")
	}

	total := 0
	summary := Summary{}
	for _, v := range summaryRows {
		if v.ColumnName == "coin" {
			total += v.Count
			if summary.Coin == nil {
				summary.Coin = make(map[string]int)
			}
			summary.Coin[v.Value] = v.Count
		}

		if v.ColumnName == "coin_ref" {
			if summary.CoinRef == nil {
				summary.CoinRef = make(map[string]int)
			}
			summary.CoinRef[v.Value] = v.Count
		}

		if v.ColumnName == "site" {
			if summary.Site == nil {
				summary.Site = make(map[string]int)
			}
			summary.Site[v.Value] = v.Count
		}

	}

	return &total, &summary, err
}

func getWhereClause(gpr *dto.GetPricesRequest) (string, []interface{}) {
	var args []interface{}
	var conditions []string

	if gpr.Coin != nil {
		conditions = append(conditions, "p.coin = ?")
		args = append(args, *gpr.Coin)
	}

	if gpr.CoinRef != nil {
		conditions = append(conditions, "p.coin_ref = ?")
		args = append(args, *gpr.CoinRef)
	}

	if gpr.Price != nil {
		conditions = append(conditions, "p.price = ?")
		args = append(args, *gpr.Price)
	}

	if gpr.StartDate != nil {
		conditions = append(conditions, "p.date >= ?")
		args = append(args, *gpr.StartDate)
	}

	if gpr.EndDate != nil {
		conditions = append(conditions, "p.date <= ?")
		args = append(args, *gpr.EndDate)
	}

	if gpr.Site != nil {
		conditions = append(conditions, "p.site = ?")
		args = append(args, *gpr.Site)
	}

	clause := ""
	if len(conditions) > 0 {
		clause = fmt.Sprintf("WHERE %s", strings.Join(conditions, " AND "))
	}
	return clause, args
}

func getJoinAndWhereClause(userId string, gpr *dto.GetPricesRequest) (string, []interface{}) {
	joinClause := `
		FROM prices AS p 
		INNER JOIN user_coins AS uc 
		ON uc.user_id = ? AND p.coin = uc.coin AND p.coin_ref = uc.coin_ref `
	whereClause, whereArgs := getWhereClause(gpr)

	joinAndWhereClause := fmt.Sprintf(`%s %s`, joinClause, whereClause)
	joinAndWhereArgs := []interface{}{userId}
	joinAndWhereArgs = append(joinAndWhereArgs, whereArgs...)

	return joinAndWhereClause, joinAndWhereArgs
}
