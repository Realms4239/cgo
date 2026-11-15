package probe

// socketToken — jeton réellement posé dans TCP_CONGESTION. Identité par
// défaut ; un hôte provisionné enregistre la v3 sous « bbr » (bbr3 n'est
// pas un token sur cet arbre) : la campagne traduit bbr3→bbr — et seulement
// là. Hors provisionné, « bbr3 » passe tel quel : le noyau le refuse
// bruyamment, jamais en silence.
var socketToken = func(cc string) string { return cc }

func SetSocketToken(f func(string) string) { socketToken = f }
