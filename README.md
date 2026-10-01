# Balanceador de carga HTTP

Projeto de estudo em Go que recebe requisições na porta `8000` e as distribui entre backends HTTP em rodízio. Usa apenas a biblioteca padrão.

O balanceador verifica a conexão com cada backend ao iniciar e a cada 20 segundos. Ele ignora os destinos indisponíveis, responde `503` quando nenhum está ativo e marca um destino como indisponível se o proxy falhar (`502` para essa requisição).

## Requisitos

- Go 1.26.1 ou superior.

## Executar

Em três terminais, a partir da raiz do projeto:

```sh
go run ./backend 8081
```

```sh
go run ./backend 8082
```

```sh
go run . http://localhost:8081 http://localhost:8082
```

Consulte o balanceador algumas vezes:

```sh
curl http://localhost:8000/api/visits
```

A resposta é um JSON como `{"instance":"8081","visits":1}`. O campo `instance` alterna entre `8081` e `8082`; cada processo mantém seu próprio contador em memória, que zera ao reiniciar.

O backend de exemplo também responde a `GET /health` com status `204`. Se nenhum endereço for passado ao balanceador, ele usa `http://localhost:8080`. Se nenhuma porta for passada ao backend, ele escuta na porta `8080`.

## Verificar

```sh
go test ./...
go vet ./...
```

## Desenvolvimento com Git

Se o projeto fosse refeito desde o início, estes commits registrariam as etapas até chegar ao estado atual:

| Etapa | Entrega | Commit |
| --- | --- | --- |
| 1 | `go.mod` | `first commit` |
| 2 | API em `backend/`, com `/health`, `/api/visits` e teste | `feat: add sample HTTP backend` |
| 3 | Proxy na porta 8000 e alternância entre destinos | `feat: balance requests across backends` |
| 4 | Verificação de saúde, tratamento de falhas e testes do balanceador | `feat: handle unavailable backends` |
