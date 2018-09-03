//
//  Group.swift
//  PrifiVPN
//
//  Created by dedis on 8/29/18.
//  Copyright © 2018 dedis. All rights reserved.
//

import Foundation

struct Group: Codable {
    let id: Int
    let name: String
    
    init(id: Int, name: String) {
        self.id = id
        self.name = name
    }
}

class GroupRepository {
    private let groupsKey = "savedGroups"
    private var groups: [Group]
    
    var activeGroupId: Int {
        get {
            return UserDefaults.standard.integer(forKey: "activeGroupId")
        }
        set {
            UserDefaults.standard.set(newValue, forKey: "activeGroupId")
        }
    }
    
    var activeGroup: Group? {
        get {
            return get(withId: activeGroupId)
        }
        set {
            activeGroupId = newValue?.id ?? 0
        }
    }
    
    var nextId: Int {
        let max = groups.compactMap { $0.id }.max() ?? 0
        return max + 1
    }
    
    init() {
        if let data = UserDefaults.standard.value(forKey: groupsKey) as? Data {
            groups = (try? PropertyListDecoder().decode(Array<Group>.self, from: data)) ?? [Group]()
        } else {
            groups = [Group]()
        }
    }
    
    func get(withId id: Int) -> Group? {
        return find(withId: id)?.first
    }
    
    func find(withId id: Int? = nil, withName name: String? = nil) -> [Group]? {
        if let id = id {
            return groups.filter { $0.id == id }
        } else if let name = name {
            return groups.filter { $0.name == name }
        }
        return nil
    }
    
    private func findIndex(_ id: Int) -> Int? {
        return groups.index { $0.id == id }
    }
    
    func add(_ group: Group) {
        groups.append(group)
        save()
    }
    
    func delete(_ group: Group) {
        delete(withId: group.id)
    }
    
    func delete(withId id: Int) {
        let index = findIndex(id)
        if let index = index {
            groups.remove(at: index)
            save()
        }
        ConfigurationRepository.remove(forGroupId: id)
    }
    
    func update(_ group: Group) {
        let index = findIndex(group.id)
        if let index = index {
            groups[index] = group
            save()
        }
    }
    
    private func save() {
        do {
            let data = try PropertyListEncoder().encode(groups)
            UserDefaults.standard.set(data, forKey: groupsKey)
        } catch {
            NSLog("Error saving Groups. \(error)")
        }
    }
    
    func all() -> [Group] {
        return groups
    }
    
    func deleteConfigurations() {
        
    }
}
