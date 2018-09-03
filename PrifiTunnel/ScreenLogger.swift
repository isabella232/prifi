//
//  ScreenLogger.swift
//  PrifiTunnel
//
//  Created on 7/11/18.

//

import PrifiMobile

class ScreenLogger: NSObject, PrifiMobilePrifiLoggerProtocol {
    var messageQueue: [String] = []
    
     func flush() -> [String] {
        let contents = Array(messageQueue)
        messageQueue.removeAll()
        return contents
    }
    
    func log(_ level: Int, msg: String!) {
        messageQueue.append(msg)
    }
    
    func close() {
        let _ = self.flush()
    }
}
