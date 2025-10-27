-- 000001_init_procedures.up.sql
-- Таблица маршрутов
-- Таблица маршрутов
IF NOT EXISTS (SELECT * FROM sysobjects WHERE name = 'ROUTE' AND xtype = 'U')
BEGIN
CREATE TABLE ROUTE (
                       ID_ROUTE BIGINT IDENTITY(1,1) PRIMARY KEY,
                       NAME NVARCHAR(255) NOT NULL UNIQUE
);
END

-- Таблица пунктов маршрута
IF NOT EXISTS (SELECT * FROM sysobjects WHERE name = 'ROUTE_ITEM' AND xtype = 'U')
BEGIN
CREATE TABLE ROUTE_ITEM (
                            ID_ROUTE_ITEM BIGINT IDENTITY(1,1) PRIMARY KEY,
                            ID_ROUTE BIGINT NOT NULL,
                            ID_CONTRACTOR_GLOBAL UNIQUEIDENTIFIER NOT NULL,
                            DISPLAY_ORDER INT NOT NULL DEFAULT 0,
                            NAME NVARCHAR(255) NOT NULL,

                            CONSTRAINT FK_ROUTE_ITEM_ROUTE FOREIGN KEY (ID_ROUTE) REFERENCES ROUTE(ID_ROUTE) ON DELETE CASCADE,
                            CONSTRAINT UQ_ROUTE_ITEM_ORDER UNIQUE (ID_ROUTE, DISPLAY_ORDER)
);
END

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
        cast(c.ID_CONTRACTOR_GLOBAL as varchar(36)) as ID_CONTRACTOR_GLOBAL,
        c.NAME,
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
    @LIMIT INT = 50,
    @NAME NVARCHAR(255) = NULL
AS
BEGIN
    SET NOCOUNT ON;

    DECLARE @OFFSET INT = (@PAGE - 1) * @LIMIT;

    -- Создаём временную таблицу для результатов
    CREATE TABLE #Results (
        ID_LOT_GLOBAL VARCHAR(36),
        LOT_NAME VARCHAR(50),
        NAME NVARCHAR(255),
        QTY FLOAT,
        PRICE_SAL MONEY,
        PRICE_PROD MONEY,
        DAYS_NO_MOVEMENT INT,
        BEST_BEFORE DATE,
        ID_GOODS_GLOBAL VARCHAR(36),
        NO_MOVE BIT,
        INTERNAL_BARCODE NVARCHAR(20)
    );

    -- Вставка данных с пагинацией
    INSERT INTO #Results
    SELECT
        L.ID_LOT_GLOBAL,
        L.LOT_NAME,
        G.NAME,
        L.QUANTITY_REM AS QTY,  -- Убрано SUM, т.к. группировка по партии
        L.PRICE_SAL,
        L.PRICE_PROD,
        ISNULL(DATEDIFF(DAY, lm_max.max_date, GETDATE()),0) AS DAYS_NO_MOVEMENT,
        ISNULL(CAST(S.BEST_BEFORE AS DATE), CAST(GETDATE() AS DATE)) AS BEST_BEFORE,
        G.ID_GOODS_GLOBAL,
        CASE
            WHEN NOT EXISTS (
                SELECT 1
                FROM LOT_MOVEMENT LMI
                WHERE LMI.CODE_OP IN (''CHEQUE'', ''INVOICE_OUT'')
                  AND LMI.ID_LOT_GLOBAL = L.ID_LOT_GLOBAL
            ) THEN 1 ELSE 0
        END AS NO_MOVE,
        L.INTERNAL_BARCODE
    FROM LOT L
    INNER JOIN STORE ST ON ST.ID_STORE = L.ID_STORE
    INNER JOIN CONTRACTOR C ON C.ID_CONTRACTOR = ST.ID_CONTRACTOR
    INNER JOIN GOODS G ON G.ID_GOODS = L.ID_GOODS
    LEFT JOIN SERIES S ON S.ID_SERIES = L.ID_SERIES
    LEFT JOIN (
        SELECT
            ID_LOT_GLOBAL,
            MAX(DATE_OP) AS max_date
        FROM LOT_MOVEMENT
        WHERE CODE_OP IN (''CHEQUE'', ''INVOICE_OUT'')
        GROUP BY ID_LOT_GLOBAL
    ) lm_max ON lm_max.ID_LOT_GLOBAL = L.ID_LOT_GLOBAL
    WHERE
        L.QUANTITY_REM > 0
        AND C.ID_CONTRACTOR_GLOBAL = @CONTRACTOR  -- ← раскомментировано!
        AND L.INCOMING_DATE < DATEADD(DAY, -@DAYS, GETDATE())
        AND (@NAME IS NULL OR (G.NAME LIKE ''%'' + @NAME + ''%'' OR L.INTERNAL_BARCODE like ''%'' + @NAME + ''%'' ))
        AND NOT EXISTS (
            SELECT 1
            FROM LOT_MOVEMENT LM
            WHERE LM.CODE_OP IN (''CHEQUE'', ''INVOICE_OUT'')
              AND LM.DATE_OP >= DATEADD(DAY, -@DAYS, GETDATE())
              AND LM.ID_LOT_GLOBAL = L.ID_LOT_GLOBAL
        )
        AND NOT EXISTS(
        SELECT NULL FROM OFFER o
        INNER JOIN OFFER_ITEM oi on o.ID_OFFER = oi.ID_OFFER
        and o.STATUS in (0,1) and oi.ID_LOT_GLOBAL = l.ID_LOT_GLOBAL)
    ORDER BY G.NAME, L.ID_LOT_GLOBAL
    OFFSET @OFFSET ROWS
    FETCH NEXT @LIMIT ROWS ONLY;

    -- Подсчёт общего количества подходящих партий (для пагинации)
    SELECT
        CEILING(COUNT(distinct l.ID_LOT_GLOBAL) * 1.0 / @LIMIT) AS TotalPages
    INTO #TotalCount
    FROM LOT L
    INNER JOIN STORE ST ON ST.ID_STORE = L.ID_STORE
    INNER JOIN CONTRACTOR C ON C.ID_CONTRACTOR = ST.ID_CONTRACTOR
    INNER JOIN GOODS G ON G.ID_GOODS = L.ID_GOODS
    WHERE
        L.QUANTITY_REM > 0
        AND C.ID_CONTRACTOR_GLOBAL = @CONTRACTOR
        AND L.INCOMING_DATE < DATEADD(DAY, -@DAYS, GETDATE())
        AND (@NAME IS NULL OR (G.NAME LIKE ''%'' + @NAME + ''%'' OR L.INTERNAL_BARCODE like ''%'' + @NAME + ''%'' ))
        AND NOT EXISTS (
            SELECT 1
            FROM LOT_MOVEMENT LM
            WHERE LM.CODE_OP IN (''CHEQUE'', ''INVOICE_OUT'')
              AND LM.DATE_OP >= DATEADD(DAY, -@DAYS, GETDATE())
              AND LM.ID_LOT_GLOBAL = L.ID_LOT_GLOBAL
        )
        AND NOT EXISTS(
        SELECT NULL FROM OFFER o
        INNER JOIN OFFER_ITEM oi on o.ID_OFFER = oi.ID_OFFER
        and o.STATUS in (0,1) and oi.ID_LOT_GLOBAL = l.ID_LOT_GLOBAL);

    -- Возврат результатов
    SELECT * FROM #Results;
    SELECT * FROM #TotalCount;

    -- Очистка (опционально, но хорошая практика)
    DROP TABLE #Results;
    DROP TABLE #TotalCount;
    END'

-- Create GetProductStockWithSalesSpeed
EXEC sp_executesql N'
CREATE PROCEDURE [dbo].[GetProductStockWithSalesSpeed]
    @DAYS INT = 30,
    @CONTRACTOR UNIQUEIDENTIFIER,
    @GOODS_ID UNIQUEIDENTIFIER = NULL,
    @SPEED_OR_ROUTE INT = 0
AS
BEGIN
    SET NOCOUNT ON;
    SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;

    DECLARE @CutoffDate DATE = DATEADD(DAY, -@DAYS, CAST(GETDATE() AS DATE));

    -- 1. Аптеки-контрагенты, удовлетворяющие условиям
    WITH eligible_contractors AS (
        SELECT DISTINCT
            C.ID_CONTRACTOR,
            C.ID_CONTRACTOR_GLOBAL,
            C.NAME
        FROM CONTRACTOR C
        INNER JOIN STORE ST ON ST.ID_CONTRACTOR = C.ID_CONTRACTOR
        WHERE
            C.ID_CONTRACTOR_GLOBAL != @CONTRACTOR
            AND (
                @SPEED_OR_ROUTE = 0
                OR EXISTS (
                    SELECT 1
                    FROM ROUTE_ITEM ri1
                    INNER JOIN ROUTE_ITEM ri2 ON ri1.ID_ROUTE = ri2.ID_ROUTE
                    WHERE ri1.ID_CONTRACTOR_GLOBAL = @CONTRACTOR
                      AND ri2.ID_CONTRACTOR_GLOBAL = C.ID_CONTRACTOR_GLOBAL
                )
            )
    ),
    -- 2. Продажи за последние @DAYS дней
    sales_agg AS (
        SELECT
            L2.ID_GOODS,
            CAST(LM.DATE_OP AS DATE) AS op_date,
            SUM(LM.QUANTITY_SUB) AS total_sold
        FROM LOT_MOVEMENT LM
        INNER JOIN LOT L2 ON L2.ID_LOT_GLOBAL = LM.ID_LOT_GLOBAL
        WHERE LM.CODE_OP IN (''CHEQUE'', ''INVOICE_OUT'')
          AND LM.DATE_OP >= @CutoffDate
        GROUP BY L2.ID_GOODS, CAST(LM.DATE_OP AS DATE)
    ),
    -- 3. Основной набор данных
    base_data AS (
        SELECT
            G.NAME AS GOOD_NAME,
            G.ID_GOODS_GLOBAL,
            EC.NAME AS CONTRACTOR_NAME,
            EC.ID_CONTRACTOR_GLOBAL,
            L.QUANTITY_REM,
            L.PRICE_SAL,
            L.PRICE_PROD,
            S.BEST_BEFORE,
            SA.total_sold,
            SA.op_date
        FROM eligible_contractors EC
        CROSS JOIN GOODS G
        INNER JOIN STORE ST ON ST.ID_CONTRACTOR = EC.ID_CONTRACTOR
        INNER JOIN LOT L ON L.ID_GOODS = G.ID_GOODS AND L.ID_STORE = ST.ID_STORE
        LEFT JOIN SERIES S ON S.ID_SERIES = L.ID_SERIES
        LEFT JOIN sales_agg SA ON SA.ID_GOODS = G.ID_GOODS
        WHERE
            (@GOODS_ID IS NULL OR G.ID_GOODS_GLOBAL = @GOODS_ID)
            AND G.ID_GOODS IS NOT NULL
    ),
    -- 4. Агрегация
    aggregated AS (
        SELECT
            GOOD_NAME,
            CAST(ID_GOODS_GLOBAL AS VARCHAR(36)) AS ID_GOODS_GLOBAL,
            CONTRACTOR_NAME,
            CAST(ID_CONTRACTOR_GLOBAL AS VARCHAR(36)) AS ID_CONTRACTOR_GLOBAL,
            SUM(CASE WHEN QUANTITY_REM > 0 THEN QUANTITY_REM ELSE 0 END) AS QTY,
            ISNULL(MAX(PRICE_SAL), 0) AS PRICE_SAL,
            ISNULL(MAX(PRICE_PROD), 0) AS PRICE_PROD,
            CONVERT(VARCHAR(10), ISNULL(BEST_BEFORE, GETDATE()), 23) AS BEST_BEFORE,
            ISNULL(SUM(total_sold), 0) AS TOTAL_SOLD_LAST_30_DAYS,
            ISNULL(ROUND(SUM(total_sold) * 1.0 / @DAYS, 2), 0) AS SALES_PER_DAY,
            ISNULL(COUNT(DISTINCT op_date), 0) AS ACTIVE_DAYS
        FROM base_data
        GROUP BY
            GOOD_NAME,
            ID_GOODS_GLOBAL,
            CONTRACTOR_NAME,
            ID_CONTRACTOR_GLOBAL,
            BEST_BEFORE
    )
    -- 5. Финальный SELECT с условным TOP и сортировкой по скорости
    SELECT
        TOP (CASE WHEN @SPEED_OR_ROUTE = 1 THEN 2147483647 ELSE 5 END)
        GOOD_NAME,
        ID_GOODS_GLOBAL,
        CONTRACTOR_NAME,
        ID_CONTRACTOR_GLOBAL,
        QTY,
        PRICE_SAL,
        PRICE_PROD,
        BEST_BEFORE,
        TOTAL_SOLD_LAST_30_DAYS,
        SALES_PER_DAY,
        ACTIVE_DAYS
    FROM aggregated
    ORDER BY SALES_PER_DAY DESC, GOOD_NAME, CONTRACTOR_NAME;
END'

IF NOT EXISTS (SELECT * FROM sysobjects WHERE name = 'OFFER' AND xtype = 'U')
BEGIN
-- Таблица заявок
CREATE TABLE OFFER (
                       ID_OFFER BIGINT IDENTITY(1,1) PRIMARY KEY,
                       NAME NVARCHAR(255) NOT NULL,
                       ID_CONTRACTOR_GLOBAL_FROM UNIQUEIDENTIFIER NOT NULL,
                       CREATED_AT DATETIME2 DEFAULT GETDATE(),
                       STATUS INT NOT NULL
);
END
IF NOT EXISTS (SELECT * FROM sysobjects WHERE name = 'OFFER_ITEM' AND xtype = 'U')
BEGIN
-- Таблица позиций заявки
CREATE TABLE OFFER_ITEM (
                            ID_OFFER_ITEM BIGINT IDENTITY(1,1) PRIMARY KEY,
                            ID_OFFER BIGINT NOT NULL,
                            ID_CONTRACTOR_GLOBAL_FROM UNIQUEIDENTIFIER NOT NULL,
                            ID_CONTRACTOR_GLOBAL_TO UNIQUEIDENTIFIER NOT NULL,
                            GOODS_ID NVARCHAR(50) NOT NULL,
                            QUANTITY INT NOT NULL,
                            ID_LOT_GLOBAL UNIQUEIDENTIFIER NOT NULL

                            CONSTRAINT FK_OFFER_ITEM_OFFER FOREIGN KEY (ID_OFFER) REFERENCES OFFER(ID_OFFER) ON DELETE CASCADE
);
END

IF OBJECT_ID('dbo.usp_GenerateInterfirmMovingFromOffer', 'P') IS NOT NULL
    DROP PROCEDURE dbo.usp_GenerateInterfirmMovingFromOffer;

EXEC sp_executesql N'
CREATE PROCEDURE [dbo].[usp_GenerateInterfirmMovingFromOffer]
    @id_offer INT
AS
BEGIN
    SET NOCOUNT ON;

    -- Шапки ПМП
    DECLARE @header TABLE (
        id_interfirm_moving_global UNIQUEIDENTIFIER,
        id_contractor_global_from UNIQUEIDENTIFIER,
        id_contractor_global_to UNIQUEIDENTIFIER
    );

    -- Склады
    DECLARE @pmp_stores TABLE (
        id_store BIGINT,
        name VARCHAR(100),
        type_store VARCHAR(10),
        id_contractor_global UNIQUEIDENTIFIER
    );

    -- Заполняем шапки
    INSERT INTO @header (id_contractor_global_from, id_contractor_global_to)
    SELECT DISTINCT
        oi.id_contractor_global_from,
        oi.id_contractor_global_to
    FROM offer_item oi
    WHERE oi.id_offer = @id_offer;

    -- Присваиваем уникальные GUID для каждой шапки
    UPDATE @header
    SET id_interfirm_moving_global = NEWID();

    -- Заполняем склады (MAIN и TRS/TRR)
    INSERT INTO @pmp_stores (id_store, name, type_store, id_contractor_global)
    SELECT
        s.id_store,
        st.NAME,
        st.MNEMOCODE,
        c.ID_CONTRACTOR_GLOBAL
    FROM contractor c
    INNER JOIN store s ON c.ID_CONTRACTOR = s.ID_CONTRACTOR
    INNER JOIN store_type st ON st.ID_STORE_TYPE_GLOBAL = s.ID_STORE_TYPE_GLOBAL
    WHERE
        (
            EXISTS (
                SELECT 1
                FROM offer_item oi
                WHERE oi.ID_CONTRACTOR_GLOBAL_TO = c.ID_CONTRACTOR_GLOBAL
                  AND oi.ID_OFFER = @id_offer
            )
            OR EXISTS (
                SELECT 1
                FROM offer_item oi
                WHERE oi.ID_CONTRACTOR_GLOBAL_FROM = c.ID_CONTRACTOR_GLOBAL
                  AND oi.ID_OFFER = @id_offer
            )
        );
        -- AND s.DATE_DELETED IS NULL; -- закомментировано, как в вашем варианте

    -- Формируем итоговый результат
    SELECT
        CAST(oi.ID_CONTRACTOR_GLOBAL_TO AS VARCHAR(36)) AS ID_CONTRACTOR_GLOBAL_TO,
        0 AS ID_INTERFIRM_MOVING,
        CAST(h.id_interfirm_moving_global AS VARCHAR(36)) AS ID_INTERFIRM_MOVING_GLOBAL,
        NULL AS MNEMOCODE,
        l.ID_STORE AS ID_STORE_FROM_MAIN,
        (SELECT p.id_store
         FROM @pmp_stores p
         WHERE p.ID_CONTRACTOR_GLOBAL = oi.ID_CONTRACTOR_GLOBAL_FROM
           AND p.type_store = ''TRS'') AS ID_STORE_FROM_TRANSIT,
        0 AS ID_CONTRACTOR_TO,
        (SELECT p.id_store
         FROM @pmp_stores p
         WHERE p.ID_CONTRACTOR_GLOBAL = oi.ID_CONTRACTOR_GLOBAL_TO
           AND p.type_store = ''MAIN'') AS ID_STORE_TO_MAIN,
        (SELECT p.id_store
         FROM @pmp_stores p
         WHERE p.ID_CONTRACTOR_GLOBAL = oi.ID_CONTRACTOR_GLOBAL_TO
           AND p.type_store = ''TRR'') AS ID_STORE_TO_TRANSIT,
        GETDATE() AS [date],
        ''SAVE'' AS DOCUMENT_STATE,
        ''Распределение неликидов заявка N '' + CAST(@id_offer AS VARCHAR(10)) AS COMMENT,
        0 AS ID_USER,
        NULL AS ID_USER2,
        0 AS SUM_SUPPLIER,
        0 AS SVAT_SUPPLIER,
        0 AS SUM_RETAIL,
        0 AS SVAT_RETAIL,
        0 AS GOODS_SENT,
        0 AS AUTH_NUM,
        NULL AS AUTH_VALID_PERIOD,
        -- Item
        0 AS ID_INTERFIRM_MOVING_ITEM,
        CAST(NEWID() AS VARCHAR(36)) AS ID_INTERFIRM_MOVING_ITEM_GLOBAL,
        oi.QUANTITY AS QUANTITY,
        l.ID_LOT AS ID_LOT_FROM,
        0 AS ID_LOT_TO,
        oi.QUANTITY * l.PRICE_SUP AS SUM_SUPPLIER,
        oi.QUANTITY * l.PVAT_SUP AS SVAT_SUPPLIER,
        l.PRICE_SAL AS PVAT_RETAIL,
        l.PVAT_SAL AS VAT_RETAIL,
        0 AS IS_WEIGHT,
        NULL AS KIZ,
        CASE WHEN l.ID_DOCUMENT_ITEM_ADD IS NOT NULL THEN 1 ELSE 0 END AS IS_KIZ
    FROM
        @header h
        INNER JOIN offer_item oi
            ON h.id_contractor_global_to = oi.ID_CONTRACTOR_GLOBAL_TO
            AND oi.ID_OFFER = @id_offer
        INNER JOIN lot l
            ON l.id_lot_global = oi.ID_LOT_GLOBAL;
END';





