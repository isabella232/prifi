//
//  GroupTableViewCell.swift
//  PrifiVPN
//
//  Copyright © 2018 dedis. All rights reserved.
//

import UIKit

protocol GroupTableViewCellDelegate {
    func nameFinishedEditing(groupId: Int, newValue: String?)
    func nameDidChange(groupId: Int, newValue: String?)
    func groupActivated(groupId: Int, activated: Bool)
}

class GroupTableViewCell: UITableViewCell, UITextFieldDelegate {
    @IBOutlet weak var nameField: UITextField!
    @IBOutlet weak var activeSwitch: UISwitch!
    
    var delegate: GroupTableViewCellDelegate?
    var groupId: Int!
    
    override func awakeFromNib() {
        super.awakeFromNib()
        // Initialization code
    }

    func bind(_ group: Group, isActive: Bool) {
        groupId = group.id
        nameField.text = group.name
        nameField.delegate = self
        nameField.addTarget(self, action: #selector(nameDidChange(_:)), for: UIControlEvents.editingChanged)
        activeSwitch.isOn = isActive
        activeSwitch.addTarget(self, action: #selector(activatedGroup(_:)), for: UIControlEvents.valueChanged)
    }
    
    override func setSelected(_ selected: Bool, animated: Bool) {
        super.setSelected(selected, animated: animated)
    }
    
    func textFieldDidEndEditing(_ textField: UITextField) {
        textField.resignFirstResponder()
        delegate?.nameFinishedEditing(groupId: groupId, newValue: textField.text)
    }
    
    @objc func nameDidChange(_ textField: UITextField) {
        delegate?.nameDidChange(groupId: groupId, newValue: textField.text)
    }
    
    @objc func activatedGroup(_ uiSwitch: UISwitch) {
        delegate?.groupActivated(groupId: groupId, activated: uiSwitch.isOn)
    }
}
