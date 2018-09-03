//
//  ServersViewController.swift
//  PrifiVPN
//
//  Copyright © 2018 dedis. All rights reserved.
//

import UIKit
import NetworkExtension

private enum Segue: String {
    case add = "addConfigSegue"
    case edit = "editConfigSegue"
}

class ConfigurationsVC: UITableViewController {
    
    var managers = [NETunnelProviderManager]()
    var targetGroup: Group!
    var editButton: UIBarButtonItem!
    
    override func viewDidLoad() {
        super.viewDidLoad()
        editButton = self.editButtonItem
        navigationItem.rightBarButtonItems?[1] = editButton
    }

    override func viewWillAppear(_ animated: Bool) {
        super.viewWillAppear(animated)
        reloadManagers()
    }
    
    override func viewWillDisappear(_ animated: Bool) {
        super.viewWillDisappear(animated)
    }
    
    func setTarget(group: Group) {
        targetGroup = group
        navigationItem.title = group.name
    }
    
    /// Re-load all of the packet tunnel configurations from the Network Extension preferences
    func reloadManagers() {
        NETunnelProviderManager.loadAllFromPreferences() { newManagers, error in
            guard let vpnManagers = newManagers else { return }
            
            self.stopObservingStatus()
            self.managers = vpnManagers.filter{ $0.groupId == self.targetGroup.id }
                .sorted(by: { (first, second) -> Bool in
                    first.priority < second.priority
                })
            self.observeStatus()
            
            self.editButton.isEnabled = self.managers.count > 0
            
            self.tableView.reloadData()
        }
    }
    
    
    /// Register for configuration change notifications.
    func observeStatus() {
        for (index, manager) in managers.enumerated() {
            NotificationCenter.default.addObserver(forName: NSNotification.Name.NEVPNStatusDidChange, object: manager.connection, queue: OperationQueue.main, using: { notification in
                self.tableView.reloadRows(at: [ IndexPath(row: index, section: 0) ], with: .fade)
            })
        }
    }
    
    /// De-register for configuration change notifications.
    func stopObservingStatus() {
        for manager in managers {
            NotificationCenter.default.removeObserver(self, name: NSNotification.Name.NEVPNStatusDidChange, object: manager.connection)
        }
    }
    
    override func tableView(_ tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
        return managers.count
    }
    
    override func tableView(_ tableView: UITableView, moveRowAt sourceIndexPath: IndexPath, to destinationIndexPath: IndexPath) {
        let source = self.managers[sourceIndexPath.row]
        managers.remove(at: sourceIndexPath.row)
        managers.insert(source, at: destinationIndexPath.row)
    }
    
    override func setEditing(_ editing: Bool, animated: Bool) {
        super.setEditing(editing, animated: animated)
        if !editing {
            updatePriorities()
        }
    }
    
    private func updatePriorities() {
        print("Saving new priorities")
        var priority = 1
        for manager in managers {
            let config = manager.toConfiguration()
            print("\(config.name ?? "") \(config.priority ?? 0) -> \(priority)")
            if config.priority != priority {
                config.priority = priority
                config.save(to: manager)
            }
            priority += 1
        }
    }
    
    override func tableView(_ tableView: UITableView, cellForRowAt indexPath: IndexPath) -> UITableViewCell {
        let cell = tableView.dequeueReusableCell(withIdentifier: "ConfigurationTableViewCell", for: indexPath)
        let manager = managers[indexPath.row]
        
        cell.textLabel?.text = manager.localizedDescription
        cell.detailTextLabel?.text = manager.protocolConfiguration?.serverAddress
        
        if manager.isEnabled {
            cell.textLabel?.font = UIFont.boldSystemFont(ofSize: 14.0)
        } else {
            cell.textLabel?.font = UIFont.systemFont(ofSize: 14.0)
        }
        
        return cell
    }
    
    override func tableView(_ tableView: UITableView, titleForHeaderInSection section: Int) -> String? {
        return "Configurations"
    }

    
    override func tableView(_ tableView: UITableView, commit editingStyle: UITableViewCellEditingStyle, forRowAt indexPath: IndexPath) {
        
        if editingStyle == .delete {
            print("Deleting at \(indexPath.row)")
            
            let manager = self.managers[indexPath.row]
            manager.removeFromPreferences { error in
                if let error = error {
                    NSLog("Failed to remove manager: \(error)")
                }
            }
            
            self.managers.remove(at: indexPath.row)
            self.tableView.deleteRows(at: [indexPath], with: .automatic)
            self.editButton.isEnabled = self.managers.count > 0
        } else if editingStyle == .insert {
            
        }
    }
    

    // MARK: - Navigation
    override func prepare(for segue: UIStoryboardSegue, sender: Any?) {
        guard let id = segue.identifier, let configurationSegue = Segue(rawValue: id) else {
            fatalError("Unknown segue performed \(segue.identifier ?? "")")
        }
        
        let dest: ConfigurationViewController?
        
        if let d = (segue.destination as? UINavigationController)?.topViewController as? ConfigurationViewController {
            dest = d
        } else if let d = segue.destination as? ConfigurationViewController {
            dest = d
        } else {
            dest = nil
        }
        
        guard let destination = dest else {
            fatalError("Unknown destination")
        }
        
        switch configurationSegue {
        case .add:
            destination.setTarget(nil, group: targetGroup, priority: getNextPriority())
        case .edit:
            if let index = tableView.indexPathForSelectedRow?.row {
                destination.setTarget(managers[index], group: targetGroup, priority: nil)
            }
        }
    }
    
    func getNextPriority() -> Int {
        let maxPriority = managers.compactMap {$0.priority}.max() ?? 0
        
        return maxPriority + 1
    }

    @IBAction func unwindCancel(segue: UIStoryboardSegue) {
        print("Canceled")
    }
    
    @IBAction func unwindSave(segue: UIStoryboardSegue) {
        print("Saved")
    }
}

extension NETunnelProviderManager {
    func toConfiguration() -> Configuration {
        return Configuration(from: self)
    }
    
    var groupId: Int {
        return toConfiguration().groupId ?? 0
    }
    
    var priority: Int {
        get {
            return toConfiguration().priority ?? 0
        }
        
        set {
            let config = toConfiguration()
            config.priority = newValue
            config.save(to: self)
        }
    }
}
