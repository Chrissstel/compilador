package codegen

import "testing"

func TestCube_Aritmeticos(t *testing.T) {
	cases := []struct {
		left, right, op string
		expected        string
	}{
		{"entero", "entero", "+", "entero"},
		{"entero", "flotante", "+", "flotante"},
		{"flotante", "entero", "*", "flotante"},
		{"flotante", "flotante", "-", "flotante"},
		{"entero", "entero", "/", "entero"},
	}
	for _, c := range cases {
		got, err := SemanticCube(c.left, c.right, c.op)
		if err != nil {
			t.Errorf("%s %s %s → error inesperado: %v", c.left, c.op, c.right, err)
		}
		if got != c.expected {
			t.Errorf("%s %s %s → esperaba %s, obtuve %s",
				c.left, c.op, c.right, c.expected, got)
		}
	}
}

func TestCube_Relacionales(t *testing.T) {
	cases := []struct {
		left, right, op string
	}{
		{"entero", "entero", ">"},
		{"entero", "flotante", "<"},
		{"flotante", "flotante", "=="},
		{"flotante", "entero", "!="},
	}
	for _, c := range cases {
		got, err := SemanticCube(c.left, c.right, c.op)
		if err != nil {
			t.Errorf("error inesperado: %v", err)
		}
		if got != "entero" {
			t.Errorf("relacional debería dar entero, obtuve %s", got)
		}
	}
}
