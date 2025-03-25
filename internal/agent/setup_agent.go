package agent

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func (p *KmAgentService) setupAgent() {

	fileData, err := os.ReadFile(AGENT_CONFIG_FILE_URI)
	if err != nil {
		fmt.Printf("failed to setup agent config caused by not able to read file : %v \n", err)
	}

	var parsedData yaml.Node
	if err := yaml.Unmarshal(fileData, &parsedData); err != nil {
		fmt.Printf("failed to setup agent config caused by not able to unmarshal config file : %v \n", err)
	}

	p.lookupAndUpdateYamlNode(&parsedData, []string{"key"}, p.AgentCfg.Key, 0)
	p.lookupAndUpdateYamlNode(&parsedData, []string{"debug"}, p.AgentCfg.debugLevel, 0)
	p.lookupAndUpdateYamlNode(&parsedData, []string{"endpoint"}, p.AgentCfg.Endpoint, 0)
	p.lookupAndUpdateYamlNode(&parsedData, []string{"interval"}, p.AgentCfg.Interval, 0)

	// creating temp file to store modified configuration
	tmpFile, err := os.CreateTemp("", "kloudmate_conf_tmp-*.yaml")
	if err != nil {
		fmt.Printf("failed to set Token : caused by not able to create temp file : %s \n", err.Error())
	}
	tmpFilePath := tmpFile.Name()

	defer os.Remove(tmpFilePath)
	defer tmpFile.Close()

	enc := yaml.NewEncoder(tmpFile)
	enc.SetIndent(2)

	if err = enc.Encode(&parsedData); err != nil {
		fmt.Printf("failed to set Token : caused by not able to encode modified config : %s \n", err.Error())
	}
	defer enc.Close()

	// saving the modified configuration
	if err = os.Rename(tmpFilePath, AGENT_CONFIG_FILE_URI); err != nil {
		fmt.Printf("failed to set Token : caused by not able to rename temp file to original config : %s \n", err.Error())
	}

}

func (p *KmAgentService) lookupAndUpdateYamlNode(node *yaml.Node, path []string, newVal string, depth int) bool {
	if depth >= len(path) {
		return false
	}
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		return p.lookupAndUpdateYamlNode(node.Content[0], path, newVal, depth)
	}
	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			kNode := node.Content[i]
			valNode := node.Content[i+1]
			if kNode.Value == path[depth] {
				if depth == len(path)-1 {
					valNode.Value = newVal
					return true
				}
				return p.lookupAndUpdateYamlNode(valNode, path, newVal, depth+1)
			}
		}
	}
	return false
}
