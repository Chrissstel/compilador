# 🦆 Compilador PATITO
 
> Compilador de un lenguaje de programación inventado llamado **PATITO**, implementado en Go. Incluye análisis léxico (lexer), análisis sintáctico (parser) con generación de AST, y análisis semántico con tabla de símbolos.
 
---
 
## ¿Qué es PATITO?
 
PATITO es un lenguaje de programación didáctico de propósito general con tipado estático, diseñado para aprender los fundamentos de la construcción de compiladores. Soporta variables de tipo entero y flotante, funciones, condicionales, ciclos y salida por pantalla.
 
Un programa válido en PATITO se ve así:
 
```
programa ejemplo ;
vars
    x, y : entero;
    resultado : flotante;
 
entero suma(a : entero, b : entero) {
    vars
        temp : entero;
    {
        temp = a + b;
    }
};
 
inicio
{
    x = 10;
    y = 20;
    si (x < y) {
        escribe("x es menor", x);
    } sino {
        escribe("y es menor o igual");
    };
    mientras (x != 0) haz {
        x = x - 1;
    };
}
fin
```
 
---
 
## Estructura del proyecto
 
```
compilador/
├── examples/               # Programas de ejemplo en PATITO
│   └── programa.patito
├── src/
│   ├── lexer/
│   │   ├── lexer.go        # Motor del análisis léxico
│   │   └── tokens.go       # Definición de tipos de tokens
│   ├── ast/
│   │   ├── ast.go          # Nodos del árbol de sintaxis abstracta
│   │   └── printer.go      # Impresión del AST en consola
│   ├── parser/
│   │   └── parser.go       # Parser recursivo descendente
│   └── semantic/
│       ├── tabla.go        # Tabla de símbolos
│       └── semantic.go     # Analizador semántico
├── testdata/
│   ├── valid/              # Programas válidos para pruebas
│   └── invalid/            # Programas inválidos para pruebas
├── main.go                 # Punto de entrada
├── go.mod
└── README.md
```
 
---
 
## Cómo correrlo
 
### Requisitos
 
- [Go](https://go.dev/dl/) `>= 1.21`
### Instalación
 
```bash
git clone https://github.com/Chrissstel/compilador.git
cd compilador
```
 
### Ejecución
 
```bash
go run main.go
```
 
Por defecto, el compilador lee el archivo `examples/programa.patito`. Para cambiar el archivo de entrada, modifica la ruta en `main.go`:
 
```go
src, err := os.ReadFile("examples/tu_programa.patito")
```
 
### Salida esperada

Al correr un programa válido verás el árbol AST impreso en consola, seguido del resultado del análisis semántico y la tabla de símbolos:

```
Programa: ejemplo
├── Vars
│   ├── Decl [x, y] : entero
│   └── Decl [resultado] : flotante
├── Funcion: suma -> entero
│   ├── Param: a : entero
│   ├── Param: b : entero
│   └── Cuerpo
│       └── Asigna: temp
│           └── Expr: a + b
└── Cuerpo
    ├── Asigna: x
    │   └── Expr: 10
    └── ...

✓ Análisis semántico correcto

╔══════════════════════════════════════╗
║        TABLA DE SÍMBOLOS             ║
╚══════════════════════════════════════╝

── Funciones ────────────────────────────
  suma(a:entero, b:entero) → entero

── Variables globales ───────────────────
  x : entero
  y : entero
  resultado : flotante

── Variables locales [suma] ─────────────
  a : entero
  b : entero
  temp : entero

────────────────────────────────────────
```

Si hay errores semánticos, se reportan todos antes de terminar:

```
=== Errores semánticos ===
SemanticError: variable 'z' no fue declarada
SemanticError: función 'miFuncion' no fue declarada
```
 
---
 
## La gramática de PATITO
 
PATITO tiene una gramática **LL(1) con lookahead de 1 token**, lo que la hace perfecta para un parser recursivo descendente. A continuación su forma EBNF (la versión compacta e implementable directamente en código):
 
```ebnf
<PROGRAMA>  →  programa id ; <V> <F> inicio <CUERPO> fin
 
<V>         →  vars (<LOOP_ID> : <TIPO> ;)+
             | ε
 
<LOOP_ID>   →  id (, id)*
 
<TIPO>      →  entero | flotante
 
<F>         →  (<DEF_FUNC> id ( (<PARAM> (, <PARAM>)*)? ) { <V> <CUERPO> } ;)*
 
<DEF_FUNC>  →  nula | entero | flotante
 
<PARAM>     →  id : <TIPO>
 
<CUERPO>    →  { <ESTATUTO>* }
 
<ESTATUTO>  →  id = <EXPRESION> ;
             | si ( <EXPRESION> ) <CUERPO> (sino <CUERPO>)? ;
             | mientras ( <EXPRESION> ) haz <CUERPO> ;
             | id ( (<EXPRESION> (, <EXPRESION>)*)? ) ;
             | escribe ( (<EXPRESION> | letrero) (, (<EXPRESION> | letrero))* ) ;
             | [ <ESTATUTO>* ]
 
<EXPRESION> →  <EXP> ((> | < | != | ==) <EXP>)?
 
<EXP>       →  <TERMINO> ((+ | -) <TERMINO>)*
 
<TERMINO>   →  <FACTOR> ((* | /) <FACTOR>)*
 
<FACTOR>    →  ( <EXPRESION> )
             | id ( (<EXPRESION> (, <EXPRESION>)*)? )
             | (+ | -)? (id | cte_ent | cte_flot)
```
 
### Palabras reservadas
 
| Keyword | Uso |
|---|---|
| `programa` | Declara el nombre del programa |
| `inicio` / `fin` | Delimitan el cuerpo principal |
| `vars` | Declara variables |
| `entero` / `flotante` | Tipos de datos |
| `nula` | Tipo de retorno vacío en funciones |
| `si` / `sino` | Condicional |
| `mientras` / `haz` | Ciclo while |
| `escribe` | Imprime en pantalla |
 
### Tipos de datos
 
- `entero` — números enteros (`42`, `-7`)
- `flotante` — números de punto flotante (`3.14`, `-0.5`)
- `letrero` — cadenas de texto solo en `escribe` (`"Hola mundo"`)
---
 
## El Lexer
 
El lexer convierte el código fuente en una secuencia plana de tokens mediante expresiones regulares. Funciona como un **autómata guiado por patrones ordenados por prioridad**:
 
1. Se itera sobre el texto restante del programa
2. Por cada posición se prueban los patrones en orden
3. El primer patrón que coincide al inicio genera un token y avanza el cursor
4. Si ningún patrón coincide, se lanza un error léxico
### Decisiones de diseño importantes
 
- Los **operadores de dos caracteres** (`==`, `!=`) están definidos **antes** que los de un carácter (`=`) para evitar que `=` consuma el primer carácter de `==`
- Las **constantes flotantes** (`\d+\.\d+`) van antes que las enteras (`\d+`) por la misma razón
- Las **palabras reservadas** usan `\b` (word boundary) y van antes que el patrón general de identificadores, así `entero` no se tokeniza como `IDENTIFICADOR`
- Los **comentarios** (`// ...`) y espacios se descartan con un `skipHandler` sin generar token
### Tokens definidos
 
| Categoría | Ejemplos |
|---|---|
| Palabras reservadas | `P_PROGRAMA`, `P_INICIO`, `P_FIN`, `P_VARS`, `P_SI`, `P_SINO`, `P_MIENTRAS`, `P_HAZ`, `P_ESCRIBE`, `P_NULA`, `P_ENTERO`, `P_FLOTANTE` |
| Identificadores | `IDENTIFICADOR` |
| Constantes | `CTE_ENTERO`, `CTE_FLOTANTE` |
| Cadenas | `LETRERO` |
| Operadores | `ASIGNACION`, `MAS`, `MENOS`, `MULTIPLICACION`, `DIVISION`, `MAYOR_QUE`, `MENOR_QUE`, `IGUAL`, `DIFERENTE` |
| Delimitadores | `ABRE_PAREN`, `CIERRA_PAREN`, `ABRE_LLAVE`, `CIERRA_LLAVE`, `ABRE_CORCHETE`, `CIERRA_CORCHETE`, `SEMICOLON`, `COMA`, `DOS_PUNTOS` |
| Fin de archivo | `EOF` |
 
---
 
## El Parser
 
El parser implementa un **análisis sintáctico recursivo descendente LL(1)**. Consume la secuencia de tokens producida por el lexer y construye un **Árbol de Sintaxis Abstracta (AST)**.
 
### Estrategia general
 
Cada regla gramatical se traduce directamente en un método de Go:
 
```
<CUERPO>  →  { <ESTATUTO>* }     →     func parseCuerpo() *ast.Cuerpo
<ASIGNA>  →  id = <EXPRESION> ;  →     func parseAsigna() *ast.Asigna
```
 
### Recursión → iteración
 
La gramática original está escrita en BNF con reglas recursivas de cola, como:
 
```
<LOOP_CUERPO> → <ESTATUTO> <LOOP_CUERPO> | ε
```
 
Go no optimiza la recursión de cola, así que estas reglas se implementan como **ciclos `for`** para evitar stack overflow:
 
```go
// En lugar de recursión:
for p.esInicioEstatuto() {
    estatutos = append(estatutos, p.parseEstatuto())
}
```
 
### Lookahead para distinguir estatutos
 
Cuando el parser ve un `IDENTIFICADOR`, necesita decidir entre una asignación y una llamada a función. Usa un **peek de un token** (LL(1)):
 
```go
case lexer.IDENTIFICADOR:
    if p.peek().Kind == lexer.ASIGNACION {
        return p.parseAsigna()   // id = ...
    }
    return p.parseLlamada()      // id( ... )
```
 
### Jerarquía de expresiones
 
Las expresiones respetan la **precedencia de operadores** mediante niveles de análisis:
 
```
EXPRESION   →  relacionales (>, <, ==, !=)   — menor precedencia
  EXP       →  aditivos    (+, -)
    TERMINO →  multiplicativos (*, /)
      FACTOR→  átomo o paréntesis             — mayor precedencia
```
 
Esto garantiza que `2 + 3 * 4` se evalúe como `2 + (3 * 4)` y no `(2 + 3) * 4`.

---

## El Análisis Semántico

Una vez generado el AST, el analizador semántico lo recorre para verificar que el programa tenga sentido más allá de su sintaxis. Esta etapa detecta errores que el parser no puede ver.

### Qué se verifica

- **Variables**: que no se declaren dos veces y que no se usen sin haber sido declaradas.
- **Funciones**: que no se declaren dos veces y que no se llamen sin haber sido declaradas.
- **Argumentos**: que el número de argumentos en cada llamada coincida con los parámetros declarados.
- **Scope**: que las variables locales de una función no sean accesibles desde fuera de ella.

### La Tabla de Símbolos

Toda la información recolectada durante el análisis se guarda en una **tabla de símbolos**, dividida en dos secciones:

- **Global**: variables declaradas en el cuerpo principal y todas las funciones del programa.
- **Local**: variables y parámetros de cada función, que solo existen dentro de su scope.

PATITO tiene un sistema de scopes de exactamente **dos niveles** — global y local de función. Los bloques `si`, `mientras` y `[ ]` no crean su propio scope, por lo que usan las variables del scope activo.

```
scope global
└── scope de función  (uno a la vez, no anidados)
```

### Errores acumulados

El analizador **no se detiene en el primer error**. Recorre el AST completo y acumula todos los errores encontrados para reportarlos juntos al final, igual que cualquier compilador real.

```go
// El analizador sigue aunque encuentre errores
func (a *Analizador) error(msg string) {
    a.errores = append(a.errores, "SemanticError: "+msg)
}
```

### Orden de análisis

El analizador sigue este orden para permitir que una función pueda llamar a otra declarada después de ella:

```
1. Registrar variables globales
2. Registrar las firmas de todas las funciones   ← primero solo el nombre y params
3. Analizar el cuerpo de cada función            ← ahora sí con el cuerpo completo
4. Analizar el cuerpo principal (inicio...fin)
```
 
Se implementaron tests automatizados en tres niveles: lexer, parser y semántico.

```bash
# correr todos los tests
go test ./...

# solo el lexer
go test ./src/lexer/

# solo el parser
go test ./src/parser/

# solo el semántico
go test ./src/semantic/
```

Para ver el detalle de cada test se añade la bandera `-v`:

```bash
go test ./... -v
```

Para ver la cobertura:

```bash
go test ./... -cover
```

Los archivos de prueba están en `testdata/`:

- `testdata/valid/` — programas que deben pasar léxico, sintáctico y semántico sin errores.
- `testdata/invalid/` — programas con errores intencionales (sintaxis rota, variables no declaradas, funciones duplicadas, argumentos incorrectos, etc.) que deben ser detectados.