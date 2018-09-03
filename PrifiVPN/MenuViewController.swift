//
//  MenuViewController.swift
//  PrifiVPN
//
//  Created on 8/10/18.
//  Copyright © 2018 dedis. All rights reserved.
//

import UIKit

private enum Segue: String {
    case groups = "groupsSegue"
    case files = "filesSegue"
    case logs = "logsSegue"
    case about = "aboutSegue"
}

class MenuViewController: UIViewController, UITableViewDataSource, UITableViewDelegate {
    @IBOutlet weak var tableView: UITableView!
    let CellIdentifier = "MenuTableViewCell"
    
    let MENU = 0
    let INFO = 1
    let HELP = 2
    
    var prifiOnly: Bool {
        get {
            return menu[0].1 as! Bool
        }
        set {
            menu[0].1 = newValue
        }
    }
    
    var menu: [(String, Any)] = [
        ("PriFi Only", false),
        ("Networks", Segue.groups)
    ]
    
    var information: [(String, String)] = [
        ("App Version", "1.3.2"),
        ("Protocol Version", "0.1")
    ]
    
    var help = [
        ("Support", "http://www.example.com")
    ]
    
    var sections = ["Status", "Information", "Help"]
    
    override func viewDidLoad() {
        super.viewDidLoad()
        tableView.dataSource = self
        tableView.delegate = self
        tableView.tableFooterView = UIView()
        
        prifiOnly = GlobalSettings.prifiOnly
        information[0].1 = getAppVersion()
    }
    
    func getAppVersion() -> String {
        guard let version = Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String else {
            return "0.1"
        }
        return version
    }
    
    override func viewWillAppear(_ animated: Bool) {
        super.viewWillAppear(animated)
    }
    
    @IBAction func onDismiss(_ sender: UIBarButtonItem) {
        dismiss(animated: true)
        let current = GlobalSettings.prifiOnly
        // Update if changed
        if current != prifiOnly {
            GlobalSettings.prifiOnly = prifiOnly
        }
    }

    @IBAction func imageClick(_ sender: UITapGestureRecognizer) {
        if let url = URL(string: "https://www.epfl.ch/") {
            UIApplication.shared.open(url, options: [:])
        }
    }
    
    func numberOfSections(in tableView: UITableView) -> Int {
        return sections.count
    }
    
    func tableView(_ tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
        if section == MENU {
            return menu.count
        } else if section == INFO {
            return information.count
        } else if section == HELP {
            return help.count
        }
        return 0
    }
    
    func tableView(_ tableView: UITableView, cellForRowAt indexPath: IndexPath) -> UITableViewCell {
        let cell = tableView.dequeueReusableCell(withIdentifier: CellIdentifier, for: indexPath)
        
        if indexPath.section == INFO {
            let (title, value) = information[indexPath.row]
            cell.accessoryType = .none
            cell.textLabel?.text = title
            cell.detailTextLabel?.text = value
            cell.detailTextLabel?.textColor = UIColor.lightGray
            
            return cell
        }
        
        if indexPath.section == HELP {
            let (title, _) = help[indexPath.row]
            cell.accessoryType = .none
            cell.textLabel?.text = title
            cell.detailTextLabel?.text = ""
            
            return cell
        }
        
        cell.detailTextLabel?.isHidden = true
        let (title, value) = menu[indexPath.row]
        cell.textLabel?.text = title
        if let value = value as? Bool {
            let valueSwitch = UISwitch()
            valueSwitch.isOn = value
            valueSwitch.addTarget(self, action: #selector(switchChanged(_:)), for: UIControlEvents.valueChanged)
            cell.accessoryView = valueSwitch
        } else {
            cell.accessoryType = UITableViewCellAccessoryType.disclosureIndicator
        }
        
        return cell
    }
    
    @objc func switchChanged(_ uiSwitch: UISwitch) {
        prifiOnly = uiSwitch.isOn
    }
    
    func tableView(_ tableView: UITableView, titleForHeaderInSection section: Int) -> String? {
        return sections[section]
    }
    
    func tableView(_ tableView: UITableView, didSelectRowAt indexPath: IndexPath) {
        if indexPath.section == HELP {
            let (_, urlString) = help[indexPath.row]
            if let url = URL(string: urlString) {
                UIApplication.shared.open(url, options: [:])
            }
        }
        
        if indexPath.section == INFO { return }
        
        let (_, value) = menu[indexPath.row]
        if let segue = value as? Segue {
            performSegue(withIdentifier: segue.rawValue, sender: self)
        }
        self.tableView.deselectRow(at: indexPath, animated: false)
    }
}
