//
//  Networking.swift
//  PrifiVPN
//
//  Created on 7/15/18.

//

import CocoaAsyncSocket
import SystemConfiguration.CaptiveNetwork
import NetworkExtension

class NetworkTools {
    static func isReachable(host: String, onPort port: UInt16,
                            withTimeout timeout: TimeInterval = TimeInterval(3),
                            delegate: GCDAsyncSocketDelegate) {
        let socket = GCDAsyncSocket(delegate: delegate, delegateQueue: DispatchQueue.main)
        do {
            NSLog("Connect to \(host):\(port)")
            try socket.connect(toHost: host, onPort: port, withTimeout: timeout)
        } catch {
            NSLog(String(describing: error))
            delegate.socketDidDisconnect?(socket, withError: error)
        }
    }
    
    static func getSsid() -> String? {
        var ssid: String?
        if let interfaces = CNCopySupportedInterfaces() as NSArray? {
            for interface in interfaces {
                if let interfaceInfo = CNCopyCurrentNetworkInfo(interface as! CFString) as NSDictionary? {
                    ssid = interfaceInfo[kCNNetworkInfoKeySSID as String] as? String
                    break
                }
            }
        }
        return ssid
    }
    
    class Request: NSObject, GCDAsyncSocketDelegate {
        let host: String
        let port: UInt16
        let timeout: TimeInterval
        
        let onCompletionHandler: (GCDAsyncSocket, String, UInt16, Error?) -> Void
        
        init(host: String, onPort port: Int, withTimeout timeout: TimeInterval = TimeInterval(3),
             onConnect onCompletionHandler: @escaping (GCDAsyncSocket, String, UInt16, Error?) -> Void) {
            
            self.host = host
            self.port = UInt16(port)
            self.timeout = timeout
            self.onCompletionHandler = onCompletionHandler
            
            super.init()
        }
        
        func socket(_ sock: GCDAsyncSocket, didConnectToHost host: String, port: UInt16) {
            onCompletionHandler(sock, host, port, nil)
        }
        
        func socketDidDisconnect(_ sock: GCDAsyncSocket, withError err: Error?) {
            if let error = err {
                onCompletionHandler(sock, host, port, error)
            }
        }
        
        // Chains another request
        func then(_ request: Request) -> Request {
            let afterCurrent = { (sock: GCDAsyncSocket, host: String, port: UInt16, error: Error?) -> Void in
                self.onCompletionHandler(sock, host, port, error)
                guard error == nil else {
                    return
                }
                request.make()
            }
            
            return Request(host: host, onPort: Int(port), withTimeout: timeout, onConnect: afterCurrent)
        }
        
        // Calls the completionHandler after the request has completed
        func then(do onCompletionHandler: @escaping (GCDAsyncSocket, String, UInt16, Error?) -> Void) -> Request {
            let afterCurrent = { (sock: GCDAsyncSocket, host: String, port: UInt16, error: Error?) -> Void in
                self.onCompletionHandler(sock, host, port, error)
                guard error == nil else {
                    onCompletionHandler(sock, host, port, error)
                    return
                }
                onCompletionHandler(sock, host, port, nil)
            }
            
            return Request(host: host, onPort: Int(port), withTimeout: timeout, onConnect: afterCurrent)
        }
        
        // Makes the request
        func make() {
            NetworkTools.isReachable(host: host, onPort: port, delegate: self)
        }
    }
}

extension NEVPNStatus: CustomStringConvertible {
    public var description: String {
        switch self {
        case .disconnected: return "Disconnected"
        case .invalid: return "Invalid"
        case .connected: return "Connected"
        case .connecting: return "Connecting"
        case .disconnecting: return "Disconnecting"
        case .reasserting: return "Reconnecting"
        }
    }
}
