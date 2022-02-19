package xmlParser

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type Node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content []byte     `xml:",innerxml"`
	Nodes   []Node     `xml:",any"`
	Module  string
}

func (x *XmlParser) nodeXml(n *Node, depth int) string {
	if len(n.Nodes) == 0 {
		return fmt.Sprintf("<%s %s/>", n.XMLName.Local, n.XmlAttributes())
	} else {
		result := fmt.Sprintf("<%s %s>", n.XMLName.Local, n.XmlAttributes())
		closing := fmt.Sprintf("</%s>", n.XMLName.Local)
		inner := ""

		for _, n := range n.Nodes {
			inner += x.nodeXml(&n, depth+1)
		}

		result = result + inner + closing
		return result
	}
}

func (n *Node) JsonAttributes() string {
	attrReprs := []string{}

	for _, a := range n.Attrs {
		attrReprs = append(attrReprs, fmt.Sprintf("%s : %s", a.Name.Local, a.Value))
	}

	if len(attrReprs) > 0 {
		return "{ " + strings.Join(attrReprs, ", ") + " }"
	} else {
		return ""
	}
}

func (n *Node) XmlAttributes() string {
	attrReprs := []string{}

	for _, a := range n.Attrs {
		value := a.Value
		value = strings.ReplaceAll(value, "&", "&amp;")
		value = strings.ReplaceAll(value, "\"", "&quot;")
		value = strings.ReplaceAll(value, "'", "&apos;")
		value = strings.ReplaceAll(value, "<", "&lt;")
		value = strings.ReplaceAll(value, ">", "&gt;")
		attrReprs = append(attrReprs, fmt.Sprintf("%s=\"%s\"", a.Name.Local, value))
	}

	if len(attrReprs) > 0 {
		return strings.Join(attrReprs, " ")
	} else {
		return ""
	}
}

func (x *XmlParser) GetXmlAttribute(attrs []xml.Attr, name string) (string, error) {
	for _, a := range attrs {
		if a.Name.Local == name {
			return a.Value, nil
		}
	}

	return "", fmt.Errorf("missing attribute %s", name)
}

func (x *XmlParser) Walk(nodes []Node, depth *int, f func(Node) (bool, error)) error {
	initialDepth := (*depth)

	for _, n := range nodes {
		(*depth) = initialDepth

		next, err := f(n)
		if err != nil {
			return err
		} else if next {
			(*depth)++
			err := x.Walk(n.Nodes, depth, f)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (x *XmlParser) NestedWalk(parent Node, f func(c Node, p Node) (bool, error)) error {

	nodes := parent.Nodes

	for _, n := range nodes {
		next, err := f(n, parent)
		if err != nil {
			return err
		} else if next {
			err := x.NestedWalk(n, f)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (x *XmlParser) PointerWalk(nodes *[]Node, depth *int, f func(*Node) (bool, error)) error {
	initialDepth := (*depth)

	for i := range *nodes {
		(*depth) = initialDepth
		n := &(*nodes)[i]

		next, err := f(n)
		if err != nil {
			return err
		} else if next {
			(*depth)++
			err := x.PointerWalk(&n.Nodes, depth, f)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (x *XmlParser) NestedPointerWalk(parent *Node, f func(c *Node, p *Node) (bool, error)) error {
	nodes := &(*parent).Nodes

	for i := range *nodes {
		n := &(*nodes)[i]

		next, err := f(n, parent)
		if err != nil {
			return err
		} else if next {
			err := x.NestedPointerWalk(n, f)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
