package codegen

import "fmt"

//recibe los dos operandos y el operador y pues da el resultado
func SemanticCube(left, right, op string) (string, error) {
	//por si acaso checamos que no entre nada raro y así poder asumir cosas mas adelante
	validTypes := map[string]bool{"entero": true, "flotante": true}
	if !validTypes[left] {
		return "", fmt.Errorf("tipo '%s' no es válido en una expresión", left)
	}
	if !validTypes[right] {
		return "", fmt.Errorf("tipo '%s' no es válido en una expresión", right)
	}

	switch op {
	case "+", "-", "*", "/":
		//si cualquiera es flotante, va a ser flotante
		if left == "flotante" || right == "flotante" {
			return "flotante", nil
		}
		return "entero", nil

	//relacionales
	//siempre es entero (nuestro booleano)
	case ">", "<", "==", "!=":
		return "entero", nil
	}

	return "", fmt.Errorf("operador '%s' no reconocido", op)
}
