//
//  GroupsVC.swift
//  PrifiVPN
//
//  Copyright © 2018 dedis. All rights reserved.
//

import UIKit

private enum Segue: String {
    case detail = "groupDetailSegue"
}

class GroupsVC: UITableViewController, GroupTableViewCellDelegate {
    // MARK: Properties
    @IBOutlet weak var addButton: UIBarButtonItem!
    
    let groupCellIdentifier = String(describing: GroupTableViewCell.self)
    
    var groupRepository: GroupRepository!
    var groupDataSource = [Group]()
    
    let sections = [
        "Networks"
    ]
    
    override func viewDidLoad() {
        super.viewDidLoad()
        navigationItem.rightBarButtonItems?[1] = editButtonItem
        groupRepository = GroupRepository()
    }
    
    override func viewWillAppear(_ animated: Bool) {
        super.viewWillAppear(animated)
        groupDataSource = groupRepository.all()
    }
    
    // MARK: - Navigation
    override func prepare(for segue: UIStoryboardSegue, sender: Any?) {
        guard let destination = segue.destination as? ConfigurationsVC else {
            return
        }
        
        guard let id = segue.identifier, let groupSegue = Segue(rawValue: id) else {
            fatalError("Unknown segue prepared")
        }
        
        switch groupSegue {
            case .detail:
            if let index = tableView.indexPathForSelectedRow?.row {
                destination.setTarget(group: groupDataSource[index])
            }
        }
    }
    
    override func numberOfSections(in tableView: UITableView) -> Int {
        return 1
    }
    
    override func tableView(_ tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
        return groupDataSource.count
    }
    
    override func tableView(_ tableView: UITableView, cellForRowAt indexPath: IndexPath) -> UITableViewCell {
        guard let cell = tableView.dequeueReusableCell(withIdentifier: groupCellIdentifier) as? GroupTableViewCell else {
            fatalError("Invalid TableViewCell")
        }
        let group = groupDataSource[indexPath.row]
        cell.bind(group, isActive: group.id == groupRepository.activeGroupId)
        cell.delegate = self
        cell.nameField.isEnabled = isEditing
        
        return cell
    }
    
    override func tableView(_ tableView: UITableView, canEditRowAt indexPath: IndexPath) -> Bool {
        return true
    }
    
    override func tableView(_ tableView: UITableView, editingStyleForRowAt indexPath: IndexPath) -> UITableViewCellEditingStyle {
        let group = groupDataSource[indexPath.row]
        if (group.id == 0) {
            return .insert
        }
        return .delete
    }
    
    override func tableView(_ tableView: UITableView, commit editingStyle: UITableViewCellEditingStyle, forRowAt indexPath: IndexPath) {
        if editingStyle == .delete {
            groupRepository.delete(groupDataSource[indexPath.row])
            groupDataSource.remove(at: indexPath.row)
            tableView.deleteRows(at: [indexPath], with: .fade)
        }
    }
    
    override func tableView(_ tableView: UITableView, titleForHeaderInSection section: Int) -> String? {
        return sections[section]
    }
    
    override func setEditing(_ editing: Bool, animated: Bool) {
        super.setEditing(editing, animated: animated)
        addButton.isEnabled = !editing
        
        if let cells = tableView.cells as? [GroupTableViewCell] {
            cells.forEach { $0.nameField.isEnabled = editing }
        }
    }
    
    @IBAction func addRow(_ sender: UIBarButtonItem) {
        let id = groupRepository.nextId
        let indexPath = IndexPath(row: groupDataSource.count, section: 0)
        
        var newName = "Network"
        
        var i = 1
        while !groupDataSource.filter({ $0.id != id }).filter({ $0.name == newName }).isEmpty {
            newName = "Network \(i)"
            i += 1
        }
        
        let group = Group(id: id, name: newName)
        
        groupRepository.add(group)
        groupDataSource.append(group)
        tableView.insertRows(at: [indexPath], with: .bottom)
        
        tableView.scrollToRow(at: indexPath, at: .top, animated: true)
        setEditing(true, animated: true)
        if let cell = tableView.cellForRow(at: indexPath) as? GroupTableViewCell {
           cell.nameField.becomeFirstResponder()
        }
    }
    
    func nameFinishedEditing(groupId: Int, newValue: String?) {
        guard let newValue = newValue, !newValue.isEmpty else {
            return
        }
        
        if let group = groupRepository.find(withId: groupId)?.first, group.name != newValue {
            print("Changed groupId: \(groupId) \(group.name) -> \(newValue)")
            let updated = Group(id: group.id, name: newValue)
            groupRepository.update(updated)
            
            let i = groupDataSource.index { $0.id == groupId }
            if let i = i {
                groupDataSource[i] = updated
            }
        }
    }
    
    func nameDidChange(groupId: Int, newValue: String?) {
        // Disables the save button if one field is empty
//        let duplicate = !groupDataSource
//            .filter{ $0.id != groupId }
//            .filter{ $0.name == newValue }.isEmpty
        
        let duplicate = false
        if newValue?.isEmpty ?? true || duplicate {
            editButtonItem.isEnabled = false
        } else {
            editButtonItem.isEnabled = true
        }
    }
    
    func groupActivated(groupId: Int, activated: Bool) {
        if activated {
            // update active group
            groupRepository.activeGroupId = groupId
            // Deactivate other switches
            if let cells = tableView.cells as? [GroupTableViewCell] {
                cells.filter{ $0.groupId != groupId }.forEach{ $0.activeSwitch.isOn = false }
            }
        } else if groupId == groupRepository.activeGroupId {
            // set to 0 if there is no active group
            groupRepository.activeGroupId = 0
        }
    }
}

extension UITableView {
    var cells:[UITableViewCell] {
        return (0..<self.numberOfSections).indices.map { (sectionIndex:Int) -> [UITableViewCell] in
            return (0..<self.numberOfRows(inSection: sectionIndex)).indices.compactMap{ (rowIndex:Int) -> UITableViewCell? in
                return self.cellForRow(at: IndexPath(row: rowIndex, section: sectionIndex))
            }
            }.flatMap{$0}
    }
}
