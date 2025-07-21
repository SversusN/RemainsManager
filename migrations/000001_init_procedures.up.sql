-- 000001_init_procedures.up.sql
IF OBJECT_ID('GetUserByUsername', 'P') IS NOT NULL
    DROP PROCEDURE GetUserByUsername;
EXEC sp_executesql N'
CREATE PROCEDURE GetUserByUsername
    @Username NVARCHAR(50)
AS
BEGIN
    SELECT PASSWORD_HASH FROM meta_user WHERE NAME = @Username
END'
IF OBJECT_ID('GetUsers', 'P') IS NOT NULL
    DROP PROCEDURE GetUsers;
EXEC sp_executesql N'
CREATE PROCEDURE GetUsers
AS
BEGIN
    SELECT name, full_name, user_num FROM meta_user
END'

EXEC sp_executesql N'
IF NOT EXISTS (SELECT * FROM sysobjects WHERE name=''contractor_routes'' AND xtype=''U'')
BEGIN
    CREATE TABLE contractor_routes (
        id_contractor_parent UNIQUEIDENTIFIER NOT NULL,
        id_contractor UNIQUEIDENTIFIER NOT NULL,

        CONSTRAINT fk_contractor_routes_parent
            FOREIGN KEY (id_contractor_parent)
            REFERENCES contractor(id_contractor_global),

        CONSTRAINT fk_contractor_routes_child
            FOREIGN KEY (id_contractor)
            REFERENCES contractor(id_contractor_global)
    );
END'

IF OBJECT_ID('GetPharmacies', 'P') IS NOT NULL
    DROP PROCEDURE GetPharmacies;
EXEC sp_executesql N'
CREATE PROCEDURE GetPharmacies
AS
BEGIN
    SELECT
        c.ID_CONTRACTOR_GLOBAL,
        c.ADDRESS,
        c.PHONE,
        c.INN
    FROM CONTRACTOR c
    INNER JOIN CONTRACTOR_2_CONTRACTOR_GROUP c2 ON c.ID_CONTRACTOR = c2.ID_CONTRACTOR
    INNER JOIN CONTRACTOR_GROUP cg ON cg.ID_CONTRACTOR_GROUP = c2.ID_CONTRACTOR_GROUP
    WHERE cg.NAME = ''Аптеки''
END'


-- Drop procedures if they exist
IF OBJECT_ID('GetInactiveStockProducts', 'P') IS NOT NULL
    DROP PROCEDURE GetInactiveStockProducts;

IF OBJECT_ID('GetProductStockWithSalesSpeed', 'P') IS NOT NULL
    DROP PROCEDURE GetProductStockWithSalesSpeed;

-- Create GetInactiveStockProducts
EXEC sp_executesql N'CREATE PROCEDURE GetInactiveStockProducts
    @DAYS INT = 30,
    @CONTRACTOR UNIQUEIDENTIFIER,
    @PAGE INT = 1,
    @LIMIT INT = 50
AS
BEGIN
    SET NOCOUNT ON;

    DECLARE @OFFSET INT = (@PAGE - 1) * @LIMIT;

    -- Создаём временную таблицу для результатов
    CREATE TABLE #Results (
                              NAME NVARCHAR(255),
                              QTY FLOAT,
                              PRICE_SAL MONEY,
                              PRICE_PROD MONEY,
                              DAYS_NO_MOVEMENT INT,
                              BEST_BEFORE DATE
    );

    INSERT INTO #Results
    SELECT
        G.NAME,
        SUM(L.QUANTITY_REM) AS QTY,
        MAX(L.PRICE_SAL) AS PRICE_SAL,
        MAX(L.PRICE_PROD) AS PRICE_PROD,
        DATEDIFF(DAY, lm_max.max_date, GETDATE()) AS DAYS_NO_MOVEMENT,
        isnull(CAST(S.BEST_BEFORE AS DATE), getdate()) AS BEST_BEFORE
    FROM LOT L
             INNER JOIN STORE ST ON ST.ID_STORE = L.ID_STORE
             INNER JOIN CONTRACTOR C ON C.ID_CONTRACTOR = ST.ID_CONTRACTOR
             INNER JOIN GOODS G ON G.ID_GOODS = L.ID_GOODS
             LEFT JOIN SERIES S ON S.ID_SERIES = L.ID_SERIES
             INNER JOIN (
        SELECT
            MAX(DATE_OP) AS max_date,
            L2.ID_GOODS
        FROM LOT_MOVEMENT LM1
                 INNER JOIN LOT L2 ON L2.ID_LOT_GLOBAL = LM1.ID_LOT_GLOBAL
        WHERE CODE_OP IN (''CHEQUE'', ''INVOICE_OUT'')
        GROUP BY L2.ID_GOODS
    ) lm_max ON lm_max.ID_GOODS = L.ID_GOODS
    WHERE QUANTITY_REM > 0
      AND C.ID_CONTRACTOR_GLOBAL = @CONTRACTOR
      AND L.INCOMING_DATE < DATEADD(DAY, -@DAYS, GETDATE())
      AND NOT EXISTS (
        SELECT 1
        FROM LOT_MOVEMENT LM
                 INNER JOIN LOT L1 ON L1.ID_LOT_GLOBAL = LM.ID_LOT_GLOBAL
        WHERE LM.CODE_OP IN (''CHEQUE'', ''INVOICE_OUT'')
          AND LM.DATE_OP BETWEEN DATEADD(DAY, -@DAYS, GETDATE()) AND GETDATE()
          AND L1.ID_GOODS = L.ID_GOODS
    )
    GROUP BY G.NAME, lm_max.max_date, CAST(S.BEST_BEFORE AS DATE)
    ORDER BY G.NAME
    OFFSET @OFFSET ROWS
        FETCH NEXT @LIMIT ROWS ONLY;


    -- Затем общее количество записей
    SELECT
        CEILING(COUNT(DISTINCT G.ID_GOODS) * 1.0 / @LIMIT) AS TotalPages    INTO #TotalCount
    FROM LOT L
             INNER JOIN STORE ST ON ST.ID_STORE = L.ID_STORE
             INNER JOIN CONTRACTOR C ON C.ID_CONTRACTOR = ST.ID_CONTRACTOR
             INNER JOIN GOODS G ON G.ID_GOODS = L.ID_GOODS
    WHERE QUANTITY_REM > 0
      AND C.ID_CONTRACTOR_GLOBAL = @CONTRACTOR
      AND L.INCOMING_DATE < DATEADD(DAY, -@DAYS, GETDATE())
      AND NOT EXISTS (
        SELECT 1
        FROM LOT_MOVEMENT LM
                 INNER JOIN LOT L1 ON L1.ID_LOT_GLOBAL = LM.ID_LOT_GLOBAL
        WHERE LM.CODE_OP IN (''CHEQUE'', ''INVOICE_OUT'')
          AND LM.DATE_OP BETWEEN DATEADD(DAY, -@DAYS, GETDATE()) AND GETDATE()
          AND L1.ID_GOODS = L.ID_GOODS
    )

    -- Сначала возвращаем данные
    SELECT * FROM #Results;
    SELECT * FROM #TotalCount;

END'

-- Create GetProductStockWithSalesSpeed
EXEC sp_executesql N'
CREATE PROCEDURE GetProductStockWithSalesSpeed
    @DAYS INT = 30,
    @CONTRACTOR UNIQUEIDENTIFIER,
    @GOODS_ID BIGINT = NULL
AS
BEGIN
    SET NOCOUNT ON;

    SELECT
        G.NAME,
        SUM(L.QUANTITY_REM) AS QTY,
        MAX(L.PRICE_SAL) AS PRICE_SAL,
        MAX(L.PRICE_PROD) AS PRICE_PROD,
        CAST(S.BEST_BEFORE AS DATE) AS BEST_BEFORE,
        ISNULL(SUM(sales.total_sold), 0) AS TOTAL_SOLD_LAST_30_DAYS,
        ISNULL(ROUND(SUM(sales.total_sold) * 1.0 / @DAYS, 2), 0) AS SALES_PER_DAY,
        ISNULL(COUNT(DISTINCT sales.op_date), 0) AS ACTIVE_DAYS
    FROM LOT L
             INNER JOIN STORE ST ON ST.ID_STORE = L.ID_STORE
             INNER JOIN CONTRACTOR C ON C.ID_CONTRACTOR = ST.ID_CONTRACTOR
             INNER JOIN GOODS G ON G.ID_GOODS = L.ID_GOODS
             LEFT JOIN SERIES S ON S.ID_SERIES = L.ID_SERIES
             LEFT JOIN (
        SELECT
            L2.ID_GOODS,
            CAST(LM.DATE_OP AS DATE) AS op_date,
            SUM(LM.QUANTITY_SUB) AS total_sold
        FROM LOT_MOVEMENT LM
                 INNER JOIN LOT L2 ON L2.ID_LOT_GLOBAL = LM.ID_LOT_GLOBAL
        WHERE LM.CODE_OP IN (''CHEQUE'', ''INVOICE_OUT'')
          AND LM.DATE_OP >= DATEADD(DAY, -@DAYS, CAST(GETDATE() AS DATE))
        GROUP BY L2.ID_GOODS, CAST(LM.DATE_OP AS DATE)
    ) AS sales ON sales.ID_GOODS = L.ID_GOODS
    WHERE L.QUANTITY_REM > 0
      AND (C.ID_CONTRACTOR_GLOBAL = @CONTRACTOR OR @CONTRACTOR IS NULL)
      AND (@GOODS_ID IS NULL OR G.ID_GOODS = @GOODS_ID)
    GROUP BY G.NAME, S.BEST_BEFORE
    ORDER BY G.NAME;
END'