//
//  LogListener.swift
//  PrifiVPN
//
//  Created on 7/9/18.

//

import PrifiMobile

class PrifiLogger: NSObject, PrifiMobilePrifiLoggerProtocol {
    func log(_ level: Int, msg: String!) {
        NSLog(msg)
    }
    
    func close() {
        
    }
}
