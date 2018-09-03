//
//  String.swift
//  PrifiVPN
//
//  Created on 7/14/18.

//

extension String {
    private static var ipv4Regex = "^(25[0-5]|2[0-4]\\d|[0-1]?\\d?\\d)(\\.(25[0-5]|2[0-4]\\d|[0-1]?\\d?\\d)){3}$"
    
    func isValidIp() -> Bool {
        return self.range(of: String.ipv4Regex, options: .regularExpression) != nil
    }
    
    func isValidPort() -> Bool {
        if let num = Int(self), 1...65535 ~= num {
            return true
        }
        return false
    }
}
