# Текстовый DSL (v1)

Документ описывает человекочитаемый DSL, поддерживаемый `dsl-runner-svc`.

Раннер поддерживает два формата:
- текстовый DSL (описан ниже);
- legacy JSON DSL (сохранен для обратной совместимости).

## 1. Форма скрипта

Каждый скрипт должен начинаться с:

```dsl
version 1
```

После этого идет один или несколько операторов.

## 2. Операторы

Поддерживаемые операторы:

```dsl
let <name> = <expr>

if <expr> {
  <stmt>...
} else {
  <stmt>...
}

repeat <expr> {
  <stmt>...
}

for <item> in <expr> {
  <stmt>...
}

for <item>, <index> in <expr> {
  <stmt>...
}

emit <expr> {
  requestMatch <expr>
  responseTemplate <expr>
  meta <expr>
}

assert <expr>, "error message"
```

Примечания:
- для `emit` обязательны все три поля: `requestMatch`, `responseTemplate`, `meta`;
- `for` внутри раннера транслируется в `forEach`;
- `assert` останавливает выполнение, если условие ложно.

## 3. Выражения

Поддерживаются:
- строки, числа, bool, null;
- object-литерал: `{ key: <expr>, ... }`;
- array-литерал: `[<expr>, ...]`;
- ссылка по пути: `input.base.request_match.path`;
- вызов функции: `randInt(1, 10)`;
- оператор конкатенации: `<expr> + <expr>`.

### 3.1 Корни путей

Разрешенные корневые пространства:
- `input.base`
- `input.params`
- `vars`
- `loop.index`
- `loop.item`

Синтаксис индекса массива:
- `input.params.items[0]`

Сокращение:
- голые идентификаторы трактуются как `vars.<name>`.
- пример: `basePath` эквивалентен `vars.basePath`.

### 3.2 Truthy/Falsy

`if` и `assert` используют правила:
- `false`, `null`, `0`, `0.0`, `""` считаются falsy;
- все остальное считается truthy.

## 4. Функции

Доступные функции:

- `get(path, default?)`
- `concat(a, b, ...)`
- `join(array, sep)`
- `lower(value)`
- `upper(value)`
- `len(value)`
- `toString(value)`
- `toInt(value)`
- `sha256(value)`
- `randInt(min, max)`
- `randBool(probability)`
- `randChoice(array)`
- `randString(alphabet, length)`
- `randUUID()`

### 4.1 Ограничения для random

Случайные функции разрешены только во время вычисления `emit`.
Использование random-функций в `let`, `if`, `repeat`, `for`, `assert` вызывает runtime-ошибку.

### 4.2 Алфавиты randString

Для `randString` поддерживаются:
- `alnum`
- `hex`
- `alpha`

## 5. Комментарии и разделители

Поддерживаемые комментарии:
- `# comment`
- `// comment`
- `/* block comment */`

Разделители операторов:
- в большинстве случаев достаточно переноса строки;
- также поддерживается `;`.

## 6. Модель ошибок (верхнеуровнево)

Возможные категории ошибок:
- ошибка парсинга;
- ошибка валидации;
- ошибка вычисления выражения (eval);
- runtime-ошибка;
- timeout;
- превышено количество сгенерированных моков;
- превышен размер результата.

## 7. Минимальный пример

```dsl
version 1

let basePath = input.base.request_match.path

repeat 2 {
  emit "gen-" + loop.index {
    requestMatch {
      method: input.base.request_match.method
      path: basePath + "/v" + loop.index
    }
    responseTemplate {
      status: 200
      body: {
        stableId: randUUID()
        n: randInt(1, 10)
      }
    }
    meta {
      from: "dsl"
      i: loop.index
    }
  }
}
```
