//
//  ConfigurationViewController.swift
//  PrifiVPN
//
//  Created on 6/26/18.
//  Copyright © 2018 dedis. All rights reserved.
//

import UIKit
import NetworkExtension
import CocoaAsyncSocket
import Eureka

private enum UnwindSegue: String {
    case save = "unwindSave"
    case cancel = "unwindCancel"
}

class ConfigurationViewController: FormViewController {
    @IBOutlet weak var saveButton: UINavigationItem!
    
    var appDelegate: AppDelegate!
    
    var config: Configuration!
    
    var targetManager: NETunnelProviderManager!
    
    var isAddMode = false
    
    override func viewDidLoad() {
        super.viewDidLoad()
        
        configureForm()
    }
    
    func setTarget(_ manager: NETunnelProviderManager?, group: Group?, priority: Int?) {
        if let manager = manager {
            // Edit
            targetManager = manager
            config = Configuration(from: manager)
            navigationItem.title = targetManager.localizedDescription
            isAddMode = false
        } else {
            guard let group = group else {
                fatalError("Group can't be nil if a configuration is being created")
            }
            guard let priority = priority else {
                fatalError("Must specify priority when creating a new configuration")
            }
            // Create
            targetManager = NETunnelProviderManager()
            config = Configuration()
            config.groupId = group.id
            config.priority = priority
            isAddMode = true
            NSLog("New Configuration with groupId: \(group.id) and priority: \(priority)")
        }
    }
    
    @IBAction func saveTarget(_ sender: Any) {
        let validationErrors = form.validate()
        guard validationErrors.isEmpty else {
            alertOk(title: "Invalid configuration", message: "Please correct your configuration.")
            return
        }
        
        
        config.save(to: targetManager) { error in
            if let error = error {
                NSLog("Failed to save configuration: \(error.localizedDescription)")
                return
            }
            self.performSegue(withIdentifier: UnwindSegue.save.rawValue, sender: sender)
        }
    }
    
    
    private func configureForm() {
        let validationOptions = ValidationOptions.validatesOnChange
        let msgRequired = "Required"
        let msgPortRange = "Must be between 0 and 65535"
        
        let onValidationError = { (cell: BaseCell, row: BaseRow) in
            if !row.isValid {
                cell.textLabel?.textColor = .red
            }
        }
        
        form +++ Section("Server Configuration")
            <<< TextRow(){ row in
                row.title = "Connection name"
                row.value = self.config?.name
                row.add(rule: RuleRequired())
            }
            .cellUpdate(onValidationError)
            .onChange({row in self.config.name = row.value})
            <<< TextRow(){ row in
                row.title = "Relay address"
                row.value = self.config?.host
                row.placeholder = "Enter server address"
                row.validationOptions = validationOptions
                row.add(rule: RuleRequired(msg: msgRequired))
                row.add(rule: RuleIP())
            }
            .cellUpdate(onValidationError)
            .onChange({row in self.config.host = row.value})
            <<< IntRow(){ row in
                row.title = "Relay port"
                row.value = self.config?.relayPort
                row.placeholder = "7000"
                row.formatter = nil
                row.add(rule: RuleRequired())
                row.add(rule: RuleGreaterThan(min: 0))
                row.add(rule: RuleSmallerThan(max: 65535))
                row.validationOptions = validationOptions
                }.cellUpdate(onValidationError).onChange({row in self.config.relayPort = row.value})
            <<< IntRow(){ row in
                row.title = "Socks port"
                row.value = self.config?.socksPort
                row.placeholder = "8090"
                row.formatter = nil
                row.add(rule: RuleRequired(msg: msgRequired))
                row.add(rule: RuleGreaterThan(min: 0, msg: msgPortRange))
                row.add(rule: RuleSmallerThan(max: 65535, msg: msgPortRange))
                row.validationOptions = validationOptions
                }.cellUpdate(onValidationError).onChange({row in self.config.socksPort = row.value})
    }
    
    override func viewWillDisappear(_ animated: Bool) {
        super.viewWillDisappear(animated)
    }
    
    override func didReceiveMemoryWarning() {
        super.didReceiveMemoryWarning()
    }
    
    override func prepare(for segue: UIStoryboardSegue, sender: Any?) {
        print(segue.identifier ?? "")
    }
    @IBAction func cancel(_ sender: Any) {
        if isAddMode {
            dismiss(animated: true, completion: nil)
        } else if let owningNavigationController = navigationController {
            owningNavigationController.popViewController(animated: true)
        }
    }
}
