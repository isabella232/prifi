//
//  PacketTunnelProvider.swift
//  PrifiTunnel
//
//  Created on 6/26/18.

//

import NetworkExtension
import PrifiMobile
import os.log

enum PrifiError: Error {
    case relayDisconnected
    case configurationError
}

class PacketTunnelProvider: NEPacketTunnelProvider {
    var session: NWUDPSession? = nil
    var conf = [String: AnyObject]()
    var screenLogger: ScreenLogger!
    var logger: PrifiLogger!

    var consoleLoggerID: Int!
    var screenLoggerID: Int!

    var prifiRunning = false
    
    var pendingStartCompletion: ((Error?) -> Void)?
    var pendingStopCompletion: (() -> Void)?
    
    override func startTunnel(options: [String : NSObject]?, completionHandler: @escaping (Error?) -> Void) {
        NSLog("Starting tunnel prifi")
        
        let config = Configuration(fromConfiguration: self.protocolConfiguration as? NETunnelProviderProtocol)
        NSLog("Config: \(config)")
        
        pendingStartCompletion = completionHandler
        
        startPrifi(config: config, autoDisconnect: config.autoDisconnect)
        configureTunnel(config: config)
    }
    
    func configureTunnel(config: Configuration) {
        let networkSettings = NEPacketTunnelNetworkSettings(tunnelRemoteAddress: "127.0.0.1")
        networkSettings.mtu = 1000
        
        let ipv4Settings = NEIPv4Settings(addresses: ["192.168.89.1"], subnetMasks: ["255.255.255.0"])
        networkSettings.ipv4Settings = ipv4Settings
        
        let proxyAddress = "127.0.0.1"
        let proxyPort = 8080
        let proxyString = "\(proxyAddress):\(proxyPort)"
        let proxySettings = NEProxySettings()
        
        proxySettings.autoProxyConfigurationEnabled = true
        proxySettings.proxyAutoConfigurationJavaScript = "" +
            "function FindProxyForURL(url, host) {" +
            "return \"SOCKS5 \(proxyString); SOCKS \(proxyString)\"; " +
        "}"
        proxySettings.matchDomains = [""]
        networkSettings.proxySettings = proxySettings
        
        setTunnelNetworkSettings(networkSettings) { (error) in
            guard error == nil else {
                print("Error configuring tunnel. ", String(describing: error))
                self.pendingStartCompletion?(error)
                self.pendingStartCompletion = nil
                return
            }
            
            print("Starting tunnel")
            self.pendingStartCompletion?(nil)
            self.pendingStartCompletion = nil
        }
    }
    
    func startPrifi(config: Configuration, autoDisconnect: Bool) {
        DispatchQueue.global(qos: .userInteractive).async {
            NSLog("Starting prifi (autodisconnect: \(autoDisconnect))")
            self.logger = PrifiLogger()
            let logging = PrifiMobileNewPrifiLogging(2, false, false, false, self.logger)
            self.consoleLoggerID = PrifiMobileRegisterPrifiLogging(logging)
            
            self.screenLogger = ScreenLogger()
            let screenLogging = PrifiMobileNewPrifiLogging(2, false, false, false, self.screenLogger)
            self.screenLoggerID = PrifiMobileRegisterPrifiLogging(screenLogging)
            
            var error = NSError?.init(nilLiteral: ())
            
            PrifiMobileSetRelayAddress(config.host, &error)
            if let error = error {
                PrifiMobileLogError("Error setting relay address. \(error)")
                self.pendingStartCompletion?(error)
            }
            
            PrifiMobileSetRelayPort(config.relayPort ?? 7000, &error)
            if let error = error {
                PrifiMobileLogError("Error setting relay port. \(error)")
                self.pendingStartCompletion?(error)
            }
            
            PrifiMobileSetRelaySocksPort(config.socksPort ?? 8090, &error)
            if let error = error {
                PrifiMobileLogError("Error setting relay socks port. \(error)")
                self.pendingStartCompletion?(error)
            }
            
            PrifiMobileSetMobileDisconnectWhenNetworkError(autoDisconnect, &error)
            if let error = error {
                PrifiMobileLogError("Error setting DisconnectWhenNetworkError. \(error)")
                self.pendingStartCompletion?(error)
            }
            
            PrifiMobileStartClient()
            
            if let pendingStart = self.pendingStartCompletion {
                NSLog("Prifi couldn't start")
                self.pendingStartCompletion = nil
                pendingStart(PrifiError.configurationError)
            } else if let pendingStop = self.pendingStopCompletion {
                NSLog("Prifi stopped manually")
                self.pendingStopCompletion = nil
                pendingStop()
            } else {
                NSLog("Prifi exited")
                self.cancelTunnelWithError(PrifiError.relayDisconnected)
            }
        }
    }
    
    func stopPrifi() {
        NSLog("Stopping Prifi VPN")
        PrifiMobileStopClient()
    }

    override func stopTunnel(with reason: NEProviderStopReason, completionHandler: @escaping () -> Void) {
        NSLog("Stop vpn tunnel. NEProviderStopReason: \(reason.description)")
        self.screenLogger = nil
        self.logger = nil
        pendingStopCompletion = completionHandler
        stopPrifi()
    }

    override func handleAppMessage(_ messageData: Data, completionHandler: ((Data?) -> Void)? = nil) {
        guard let handler = completionHandler else {
            return
        }

        let response = NSKeyedArchiver.archivedData(withRootObject: screenLogger.flush())
        handler(response)
    }

    override func sleep(completionHandler: @escaping () -> Void) {
        PrifiMobileLogInfo("Device will sleep")
        completionHandler()
    }

    override func wake() {
        PrifiMobileLogInfo("Device will wake")
    }
}
