package graph

var EntityTypesCache = []EntityType{
	{
		Name:    "diagnostics",
		Aliases: []string{"diagnóstico", "avaliado"},
	},
	{
		Name:    "farms",
		Aliases: []string{"fazenda", "propriedade", "área"},
	},
	{
		Name:    "companies",
		Aliases: []string{"empresa", "companhia", "cliente"},
	},
	{
		Name:    "addresses",
		Aliases: []string{"endereço", "cidade", "estado"},
	},
	{
		Name:    "users",
		Aliases: []string{"usuário", "pessoa", "nome do usuário"},
	},
	{
		Name:    "analysts",
		Aliases: []string{"analista", "especialista", "técnico"},
	},
	{
		Name:    "checklists",
		Aliases: []string{"checklist", "visita", "verificação"},
	},
	{
		Name:    "scores",
		Aliases: []string{"pontuação", "nota", "resultado", "score"},
	},
}
