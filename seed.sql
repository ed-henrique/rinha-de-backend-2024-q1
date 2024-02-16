CREATE TABLE IF NOT EXISTS CLIENTE (ID INTEGER, LIMITE INTEGER, SALDO INTEGER CHECK(SALDO >= -LIMITE));
CREATE TABLE IF NOT EXISTS TRANSACAO (VALOR INTEGER, TIPO TEXT, DESCRICAO TEXT, CLIENTE INTEGER, REALIZADA_EM TEXT);

INSERT INTO CLIENTE (ID, LIMITE, SALDO) VALUES (1, 100000, 0);
INSERT INTO CLIENTE (ID, LIMITE, SALDO) VALUES (2, 80000, 0);
INSERT INTO CLIENTE (ID, LIMITE, SALDO) VALUES (3, 1000000, 0);
INSERT INTO CLIENTE (ID, LIMITE, SALDO) VALUES (4, 10000000, 0);
INSERT INTO CLIENTE (ID, LIMITE, SALDO) VALUES (5, 500000, 0);
