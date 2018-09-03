//
//  Configuration.swift
//  PrifiVPN
//
//  Copyright © 2018 dedis. All rights reserved.
//

import Foundation
import NetworkExtension

class Configuration: CustomStringConvertible {
    var name: String?
    var host: String?
    var relayPort: Int?
    var socksPort: Int?
    var groupId: Int?
    var priority: Int?
    var autoDisconnect: Bool
    
    public var description: String {
        return "Configuration(host: \(String(describing: host)), relayPort: \(String(describing: relayPort)), socksPort: \(String(describing: socksPort)), groupId: \(String(describing: groupId)), priority: \(String(describing: priority)), autoDisconnect: \(autoDisconnect)"
    }
    
    init(name: String?, host: String?, relayPort: Int?, socksPort: Int?, groupId: Int?, priority: Int?, autoDisconnect: Bool = false) {
        self.name = name
        self.host = host
        self.relayPort = relayPort
        self.socksPort = socksPort
        self.groupId = groupId
        self.priority = priority
        self.autoDisconnect = autoDisconnect
    }
    
    convenience init(name: String?, host: String?) {
        self.init(name: name, host: host, relayPort: 7000, socksPort: 8090, groupId: 0, priority: 0)
    }
    
    convenience init() {
        self.init(name: "", host: "")
    }
    
    convenience init(from manager: NETunnelProviderManager) {
        let configuration = manager.protocolConfiguration as? NETunnelProviderProtocol
        
        self.init(fromConfiguration: configuration)
        self.name = manager.localizedDescription
    }
    
    convenience init(fromConfiguration configuration: NETunnelProviderProtocol?) {
        self.init(name: "",
                  host: configuration?.serverAddress,
                  relayPort: configuration?.providerConfiguration?["relayPort"] as? Int,
                  socksPort: configuration?.providerConfiguration?["socksPort"] as? Int,
                  groupId: configuration?.providerConfiguration?["groupId"] as? Int,
                  priority: configuration?.providerConfiguration?["priority"] as? Int,
                  autoDisconnect: configuration?.providerConfiguration?["autoDisconnect"] as? Bool ?? false
        )
    }
    
    func save(to manager: NETunnelProviderManager, completionHandler: ((Error?) -> Void)? = nil) {
        manager.localizedDescription = name
        
        let configuration = manager.protocolConfiguration as? NETunnelProviderProtocol ?? NETunnelProviderProtocol()
        
        configuration.serverAddress = host
        configuration.providerConfiguration = [
            "relayPort": relayPort ?? 0,
            "socksPort": socksPort ?? 0,
            "groupId": groupId ?? 0,
            "priority": priority ?? 0,
            "autoDisconnect": autoDisconnect
        ]
        
        manager.protocolConfiguration = configuration
        
        manager.saveToPreferences(completionHandler: completionHandler)
    }
}

class ConfigurationRepository {
    static func remove(forGroupId id: Int, completionHandler: (([NETunnelProviderManager]?) -> Error?)? = nil) {
        NETunnelProviderManager.loadAllFromPreferences { managers, error in
            guard let managers = managers else { return }
            if let error = error { NSLog(error.localizedDescription)}
            for manager in managers {
                let config = Configuration(from: manager)
                if let groupId = config.groupId, groupId == id {
                    manager.removeFromPreferences()
                }
            }
        }
    }
}
