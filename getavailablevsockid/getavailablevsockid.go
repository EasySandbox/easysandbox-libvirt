package getavailablevsockid

import (
	"encoding/xml"
	"fmt"

	"libvirt.org/go/libvirt"
)

// domain.devices.vsock.cid.address
type vsockXML struct {
	XMLName xml.Name `xml:"domain"`
	Devices struct {
		XMLName xml.Name `xml:"devices"`
		Vsock   struct {
			XMLName xml.Name `xml:"vsock"`
			Cid     struct {
				XMLName xml.Name `xml:"cid"`
				Address int   `xml:"address,attr"`
			} `xml:"cid"`
		} `xml:"vsock"`
	} `xml:"devices"`
}

func GetAvailableVSockID(libvirtConn *libvirt.Connect) (int, error) {

	domainIDs, listDomainsErr := libvirtConn.ListDomains()
	if listDomainsErr != nil {
		return 0, listDomainsErr
	}

	// iterate over all domain IDs and check if the vsock ID is already in use
	for _, domainID := range domainIDs {
		domain, lookupDomainErr := libvirtConn.LookupDomainById(domainID)
		if lookupDomainErr != nil {
			return 0, lookupDomainErr
		}
		domainXMLString, err := domain.GetXMLDesc(0)
		if err != nil {
			return 0, err
		}
		//fmt.Println(domainXMLString)
		var vsockXMLData vsockXML

		err = xml.Unmarshal([]byte(domainXMLString), &vsockXMLData)
		if err != nil {
			return 0, err
		}
		domainName, getNameErr := domain.GetName()
		if getNameErr != nil {
			return 0, getNameErr
		}
		fmt.Println(domainName, vsockXMLData.Devices.Vsock.Cid.Address)

	}
	return 1, nil

}
