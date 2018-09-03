//
//  PrifiVpn.swift
//  PrifiVPN
//
//  Copyright © 2018 dedis. All rights reserved.

import NetworkExtension
import CocoaAsyncSocket

class PrifiVpn {
    var manager = NETunnelProviderManager()
    // Hard code VPN configurations
    let tunnelBundleId = "com.lca.PrifiVPN.PrifiTunnel"
    let serverAddress = "192.168.0.13" // Dummy address
    let serverPort = "8080"
    let mtu = "1400"
    let ip = "10.8.0.2"
    let subnet = "255.255.255.0"
    let dns = "8.8.8.8,8.4.4.4"

    func initVPNTunnelProviderManager(completionHandler: @escaping (NETunnelProviderManager?, Error?) -> Void) {
        NSLog("Starting tunnel provider")
        NETunnelProviderManager.loadAllFromPreferences { (savedManagers: [NETunnelProviderManager]?, error: Error?) in
            if let error = error {
                completionHandler(nil, error)
            }
            if let savedManagers = savedManagers {
                if savedManagers.count > 0 {
                    self.manager = savedManagers[0]
                }
            }

            self.manager.loadFromPreferences(completionHandler: { (error:Error?) in
                if let error = error {
                    print(error)
                }

                let providerProtocol = NETunnelProviderProtocol()
                providerProtocol.providerBundleIdentifier = self.tunnelBundleId

                providerProtocol.providerConfiguration = ["port": self.serverPort,
                                                          "server": self.serverAddress,
                                                          "ip": self.ip,
                                                          "subnet": self.subnet,
                                                          "mtu": self.mtu,
                                                          "dns": self.dns
                ]
                providerProtocol.serverAddress = self.serverAddress
                self.manager.protocolConfiguration = providerProtocol
                self.manager.localizedDescription = "Private WiFi"
                self.manager.isEnabled = true

                self.manager.saveToPreferences(completionHandler: { (error:Error?) in
                    if let error = error {
                        completionHandler(nil, error)
                    } else {
                        completionHandler(self.manager, nil)
                        print("Save successfully")
                    }
                })
            })
        }
    }
}
