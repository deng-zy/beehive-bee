package bee

import "github.com/bwmarrin/snowflake"

// newSnowflake new snowflake instance
func newSnowflake(node, epoch int64) (*snowflake.Node, error) {
	snowflake.Epoch = epoch
	return snowflake.NewNode(node)
}
