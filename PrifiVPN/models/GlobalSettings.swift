//
//  GlobalSettings.swift
//  PrifiVPN
//
//  Created by DeDiS on 9/1/18.
//  Copyright © 2018 dedis. All rights reserved.
//

import Foundation

class GlobalSettings {
    static var prifiOnly : Bool {
        get {
            return UserDefaults.standard.object(forKey: "prifiOnly") as? Bool ?? true
        }
        set {
            UserDefaults.standard.set(newValue, forKey: "prifiOnly")
        }
    }
    
    static var firstInit : Bool {
        get {
            return UserDefaults.standard.object(forKey: "firstInit") as? Bool ?? true
        }
        set {
            UserDefaults.standard.set(newValue, forKey: "firstInit")
        }
    }
}
