# 🦆 Compilador PATITO
 
> Compilador de un lenguaje de programación inventado llamado **PATITO**, implementado en Go. Incluye análisis léxico (lexer) y análisis sintáctico (parser) con generación de AST.
 
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
│   └── parser/
│       └── parser.go       # Parser recursivo descendente
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
 
Al correr un programa válido verás el árbol AST impreso en consola:
 
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

### Casos de prueba
 
Se implementaron tests divididos en dos niveles: tests del lexer y tests del parser.
Cómo correr los tests
 
```bash
#para correr todos los tests
go test ./...

#para correr solo los del lexer
go test ./src/lexer/

#para correr solo los del parser
go test ./src/parser/

```
 
Para ver el detalle de cada test, se puede añadir la bandera “-v” y se va a imprimir el nombre y el resultado de cada caso. (Por ejemplo: “ go test ./... -v ”.
 
---
 
## Estado actual del compilador
 
- [x] Análisis léxico (Lexer)
- [x] Análisis sintáctico (Parser)
- [x] Generación de AST
- [x] Impresión del AST
- [ ] Tabla de símbolos
- [ ] Análisis semántico
- [ ] Generación de código intermedio
- [ ] Ejecución / Máquina virtual
---
 
## Tecnologías
 
- **Lenguaje**: [Go](https://go.dev/) 1.21+
- **Sin dependencias externas** — implementación desde cero
---
 
*Proyecto académico — Construcción de Compiladores*
