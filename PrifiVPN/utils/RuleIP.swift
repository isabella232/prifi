//
//  RuleIP.swift
//  PrifiVPN
//
//  Copyright © 2018 dedis. All rights reserved.
//

import Eureka
import Foundation

public class RuleIP: RuleRegExp {
    static let IPv4 = "^(25[0-5]|2[0-4]\\d|[0-1]?\\d?\\d)(\\.(25[0-5]|2[0-4]\\d|[0-1]?\\d?\\d)){3}$"
    
    public init(msg: String = "Field value should be a IP!", id: String? = nil) {
        super.init(regExpr: RuleIP.IPv4, allowsEmpty: true, msg: msg, id: id)
    }
    
}
